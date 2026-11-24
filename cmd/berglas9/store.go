package main

import (
	"context"
	"io"
	"time"

	"github.com/GoogleCloudPlatform/berglas/pkg/berglas"
)

type secretRef struct {
	Name           string
	Generation     int64
	Metageneration int64
	UpdatedAt      time.Time
}

type secretStore interface {
	List(ctx context.Context) ([]secretRef, error)
	Access(ctx context.Context, name string, generation int64) ([]byte, error)
	Create(ctx context.Context, name string, plaintext []byte) error
	Update(ctx context.Context, ref secretRef, plaintext []byte) error
}

type berglasStore struct {
	client *berglas.Client
	bucket string
	key    string
}

func newBerglasStore(bucket, key string) (*berglasStore, error) {
	ctx := context.Background()
	id, err := loadGcloudIdentity(ctx)
	if err != nil {
		return nil, err
	}
	client, err := berglas.New(ctx, id.clientOptions()...)
	if err != nil {
		return nil, err
	}
	client.SetLogOutput(io.Discard)
	return &berglasStore{client: client, bucket: bucket, key: key}, nil
}

func (s *berglasStore) List(ctx context.Context) ([]secretRef, error) {
	resp, err := s.client.List(ctx, &berglas.StorageListRequest{Bucket: s.bucket})
	if err != nil {
		return nil, err
	}
	out := make([]secretRef, 0, len(resp.Secrets))
	for _, sec := range resp.Secrets {
		if sec == nil {
			continue
		}
		out = append(out, secretRef{
			Name:           sec.Name,
			Generation:     sec.Generation,
			Metageneration: sec.Metageneration,
			UpdatedAt:      sec.UpdatedAt,
		})
	}
	return out, nil
}

func (s *berglasStore) Access(ctx context.Context, name string, generation int64) ([]byte, error) {
	return s.client.Access(ctx, &berglas.StorageAccessRequest{
		Bucket:     s.bucket,
		Object:     name,
		Generation: generation,
	})
}

func (s *berglasStore) Create(ctx context.Context, name string, plaintext []byte) error {
	_, err := s.client.Create(ctx, &berglas.StorageCreateRequest{
		Bucket:    s.bucket,
		Object:    name,
		Key:       s.key,
		Plaintext: plaintext,
	})
	return err
}

func (s *berglasStore) Update(ctx context.Context, ref secretRef, plaintext []byte) error {
	_, err := s.client.Update(ctx, &berglas.StorageUpdateRequest{
		Bucket:         s.bucket,
		Object:         ref.Name,
		Generation:     ref.Generation,
		Metageneration: ref.Metageneration,
		Key:            s.key,
		Plaintext:      plaintext,
	})
	return err
}
