package infra

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	herenowv1 "github.com/agarwalvivek29/here.now/packages/schema/generated/go/herenow/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// FileStore is a file-backed metadata + audit store for single-binary local
// deploys. State is held in memory and persisted as protojson; the audit log is
// an append-only, hash-chained JSONL file. Not for concurrent multi-process use —
// Postgres is that adapter.
type FileStore struct {
	mu     sync.Mutex
	dir    string
	arts   map[string]*herenowv1.Artifact
	grants []*herenowv1.Grant
	last   string // chain head
	seq    int64
}

func NewFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	s := &FileStore{dir: dir, arts: map[string]*herenowv1.Artifact{}}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileStore) artsPath() string   { return filepath.Join(s.dir, "artifacts.json") }
func (s *FileStore) grantsPath() string { return filepath.Join(s.dir, "grants.json") }
func (s *FileStore) auditPath() string  { return filepath.Join(s.dir, "audit.log") }

func (s *FileStore) load() error {
	if b, err := os.ReadFile(s.artsPath()); err == nil {
		raw := map[string]json.RawMessage{}
		if err := json.Unmarshal(b, &raw); err != nil {
			return err
		}
		for slug, msg := range raw {
			a := &herenowv1.Artifact{}
			if err := protojson.Unmarshal(msg, a); err != nil {
				return err
			}
			s.arts[slug] = a
		}
	}
	if b, err := os.ReadFile(s.grantsPath()); err == nil {
		var raw []json.RawMessage
		if err := json.Unmarshal(b, &raw); err != nil {
			return err
		}
		for _, msg := range raw {
			g := &herenowv1.Grant{}
			if err := protojson.Unmarshal(msg, g); err != nil {
				return err
			}
			s.grants = append(s.grants, g)
		}
	}
	if b, err := os.ReadFile(s.auditPath()); err == nil {
		for _, line := range bytes.Split(b, []byte("\n")) {
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}
			ev := &herenowv1.AuditEvent{}
			if err := protojson.Unmarshal(line, ev); err != nil {
				return err
			}
			s.last = ev.GetHash()
			s.seq = ev.GetSeq()
		}
	}
	return nil
}

func (s *FileStore) PutArtifact(a *herenowv1.Artifact) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.arts[a.GetSlug()] = a
	return s.persistArtifacts()
}

func (s *FileStore) persistArtifacts() error {
	out := map[string]json.RawMessage{}
	for slug, a := range s.arts {
		b, err := protojson.Marshal(a)
		if err != nil {
			return err
		}
		out[slug] = b
	}
	return writeJSON(s.artsPath(), out)
}

func (s *FileStore) GetArtifact(slug string) (*herenowv1.Artifact, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.arts[slug]
	return a, ok, nil
}

func (s *FileStore) ListByOwner(sub string) ([]*herenowv1.Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*herenowv1.Artifact
	for _, a := range s.arts {
		if a.GetOwnerSub() == sub {
			out = append(out, a)
		}
	}
	return out, nil
}

// ListByGrantee returns the deduped set of artifacts shared with sub via a
// Grant. It joins grants→artifacts: a slug is included only if a grant names
// sub as grantee AND the artifact still exists. This is strictly grant-based
// "shared with me" — the caller's own artifacts are NOT included implicitly.
func (s *FileStore) ListByGrantee(sub string) ([]*herenowv1.Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*herenowv1.Artifact
	seen := map[string]struct{}{}
	for _, g := range s.grants {
		if g.GetGranteeSub() != sub {
			continue
		}
		slug := g.GetSlug()
		if _, dup := seen[slug]; dup {
			continue
		}
		a, ok := s.arts[slug]
		if !ok {
			continue // dangling grant — artifact deleted
		}
		seen[slug] = struct{}{}
		out = append(out, a)
	}
	return out, nil
}

// ListByVisibility returns every artifact whose visibility equals v. It backs
// the dashboard's Org tab (ListByVisibility(ORG)); the caller is responsible for
// any further filtering (e.g. excluding its own artifacts).
func (s *FileStore) ListByVisibility(v herenowv1.Visibility) ([]*herenowv1.Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*herenowv1.Artifact
	for _, a := range s.arts {
		if a.GetVisibility() == v {
			out = append(out, a)
		}
	}
	return out, nil
}

func (s *FileStore) AddGrant(g *herenowv1.Grant) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.grants = append(s.grants, g)
	out := make([]json.RawMessage, 0, len(s.grants))
	for _, gr := range s.grants {
		b, err := protojson.Marshal(gr)
		if err != nil {
			return err
		}
		out = append(out, b)
	}
	return writeJSON(s.grantsPath(), out)
}

func (s *FileStore) Grants(slug string) ([]*herenowv1.Grant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*herenowv1.Grant
	for _, g := range s.grants {
		if g.GetSlug() == slug {
			out = append(out, g)
		}
	}
	return out, nil
}

// Append writes one hash-chained, append-only audit row.
func (s *FileStore) Append(ev *herenowv1.AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	ev.Seq = s.seq
	ev.PrevHash = s.last
	ev.Hash = hashEvent(ev)
	s.last = ev.GetHash()

	line, err := protojson.Marshal(ev)
	if err != nil {
		return err
	}
	fh, err := os.OpenFile(s.auditPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer fh.Close()
	_, err = fh.Write(append(line, '\n'))
	return err
}

func hashEvent(ev *herenowv1.AuditEvent) string {
	payload := fmt.Sprintf("%d|%s|%s|%s|%s|%t|%s",
		ev.GetSeq(), ev.GetTs().AsTime().UTC().Format("2006-01-02T15:04:05.000000000Z07:00"),
		ev.GetPrincipalSub(), ev.GetSlug(), ev.GetAction().String(), ev.GetAllowed(), ev.GetPrevHash())
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

// writeJSON atomically replaces path with the JSON encoding of v. It writes to a
// temp file in the same directory (so os.Rename stays atomic on one filesystem),
// then renames over the target. A crash leaves either the old or new file intact,
// never a partial write.
func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	dir, base := filepath.Split(path)
	tmp, err := os.CreateTemp(dir, base+".tmp-*")
	if err != nil {
		return err
	}
	// Clean up the temp file on any error before the rename succeeds.
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
