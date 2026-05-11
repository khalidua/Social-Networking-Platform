package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"social-networking-platform/posts-service/internal/apiresponse"
	"social-networking-platform/posts-service/internal/apperrors"
	"social-networking-platform/posts-service/internal/domain"
	handlers "social-networking-platform/posts-service/internal/handler/http"
	postkafka "social-networking-platform/posts-service/internal/repository/kafka"
	pgrepo "social-networking-platform/posts-service/internal/repository/postgres"
	"social-networking-platform/posts-service/internal/service"
	httptransport "social-networking-platform/posts-service/internal/transport/http"
)

type recordPostProducer struct {
	events       []domain.Post
	interactions []domain.PostInteraction
}

func (p *recordPostProducer) PublishCreated(ctx context.Context, post domain.Post) error {
	p.events = append(p.events, post)
	return nil
}

func (p *recordPostProducer) PublishInteracted(ctx context.Context, interaction domain.PostInteraction) error {
	p.interactions = append(p.interactions, interaction)
	return nil
}

func (p *recordPostProducer) Close() error { return nil }

func stackRouter(tb testing.TB, repo pgrepo.PostRepository, producer postkafka.PostProducer) http.Handler {
	tb.Helper()
	svc := service.NewPostService(repo, producer)
	h := handlers.NewPostHandler(svc)
	return httptransport.NewRouter("posts-integration", h)
}

func request(method string, path string, body string, userID string) *http.Request {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	r := httptest.NewRequest(method, path, reader)
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	r.Header.Set("X-Request-ID", "int-req")
	if userID != "" {
		r.Header.Set("X-User-ID", userID)
	}
	return r
}

func decodeSuccessPost(t *testing.T, rec *httptest.ResponseRecorder) domain.Post {
	t.Helper()
	var env apiresponse.SuccessEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("json.Unmarshal success envelope: %v", err)
	}
	if !env.Success {
		t.Fatalf("expected success envelope, got %s", rec.Body.String())
	}

	raw, err := json.Marshal(env.Data)
	if err != nil {
		t.Fatalf("json.Marshal nested data: %v", err)
	}

	var post domain.Post
	if err := json.Unmarshal(raw, &post); err != nil {
		t.Fatalf("json.Unmarshal post: %v", err)
	}
	return post
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) apiresponse.ErrorEnvelope {
	t.Helper()
	var env apiresponse.ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("json.Unmarshal error envelope: %v", err)
	}
	return env
}

func TestIntegration_PostLifecycleAndPublish(t *testing.T) {
	repo := pgrepo.NewInMemoryPostRepository()
	producer := &recordPostProducer{}
	h := stackRouter(t, repo, producer)

	rec := httptest.NewRecorder()
	r := request(http.MethodPost, "/api/v1/posts", `{"content":"hello world"}`, "user:alice")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Create expected 201 got %d %s", rec.Code, rec.Body.String())
	}
	created := decodeSuccessPost(t, rec)
	if created.ID == "" {
		t.Fatal("expected created post id")
	}
	if created.AuthorID != "user:alice" || created.Content != "hello world" {
		t.Fatalf("unexpected created post: %+v", created)
	}
	if len(producer.events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(producer.events))
	}
	if producer.events[0].ID != created.ID || producer.events[0].AuthorID != "user:alice" {
		t.Fatalf("unexpected published event: %+v", producer.events[0])
	}

	rec = httptest.NewRecorder()
	r = request(http.MethodGet, "/api/v1/posts/"+created.ID, "", "")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("Get expected 200 got %d %s", rec.Code, rec.Body.String())
	}
	read := decodeSuccessPost(t, rec)
	if read.ID != created.ID || read.Content != "hello world" {
		t.Fatalf("unexpected read post: %+v", read)
	}

	rec = httptest.NewRecorder()
	r = request(http.MethodGet, "/api/v1/posts?authorId=user:alice", "", "")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("List expected 200 got %d %s", rec.Code, rec.Body.String())
	}
	var listEnv apiresponse.SuccessEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &listEnv); err != nil {
		t.Fatalf("json.Unmarshal list envelope: %v", err)
	}
	rawList, err := json.Marshal(listEnv.Data)
	if err != nil {
		t.Fatalf("json.Marshal list data: %v", err)
	}
	var posts []domain.Post
	if err := json.Unmarshal(rawList, &posts); err != nil {
		t.Fatalf("json.Unmarshal list posts: %v", err)
	}
	if len(posts) != 1 || posts[0].ID != created.ID {
		t.Fatalf("unexpected listed posts: %+v", posts)
	}

	rec = httptest.NewRecorder()
	r = request(http.MethodPut, "/api/v1/posts/"+created.ID, `{"content":"edited"}`, "user:bob")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Unauthorized Update expected 403 got %d %s", rec.Code, rec.Body.String())
	}
	errEnv := decodeError(t, rec)
	if errEnv.Error != apperrors.CodeForbidden {
		t.Fatalf("expected forbidden code, got %+v", errEnv)
	}

	rec = httptest.NewRecorder()
	r = request(http.MethodPut, "/api/v1/posts/"+created.ID, `{"content":"edited"}`, "user:alice")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("Authorized Update expected 200 got %d %s", rec.Code, rec.Body.String())
	}
	updated := decodeSuccessPost(t, rec)
	if updated.Content != "edited" {
		t.Fatalf("unexpected updated post: %+v", updated)
	}
	if len(producer.events) != 1 {
		t.Fatalf("expected create event to publish only once, got %d", len(producer.events))
	}

	rec = httptest.NewRecorder()
	r = request(http.MethodPost, "/api/v1/posts/"+created.ID+"/interactions", `{"interaction_type":"like"}`, "user:bob")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("Interact expected 202 got %d %s", rec.Code, rec.Body.String())
	}
	if len(producer.interactions) != 1 {
		t.Fatalf("expected 1 published interaction, got %d", len(producer.interactions))
	}
	if producer.interactions[0].PostID != created.ID ||
		producer.interactions[0].PostAuthorID != "user:alice" ||
		producer.interactions[0].ActorID != "user:bob" ||
		producer.interactions[0].InteractionType != "like" {
		t.Fatalf("unexpected published interaction: %+v", producer.interactions[0])
	}

	rec = httptest.NewRecorder()
	r = request(http.MethodDelete, "/api/v1/posts/"+created.ID, "", "user:alice")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Delete expected 204 got %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	r = request(http.MethodGet, "/api/v1/posts/"+created.ID, "", "")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("Get after delete expected 404 got %d %s", rec.Code, rec.Body.String())
	}
	errEnv = decodeError(t, rec)
	if errEnv.Error != apperrors.CodeNotFound {
		t.Fatalf("expected not found code, got %+v", errEnv)
	}
}
