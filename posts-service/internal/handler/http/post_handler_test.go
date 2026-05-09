package handlers

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
	"social-networking-platform/posts-service/internal/middleware"
	"social-networking-platform/posts-service/internal/service"
)

type mockPostService struct {
	createPostFunc        func(ctx context.Context, authorID string, content string) (*domain.Post, error)
	getPostFunc           func(ctx context.Context, id string) (*domain.Post, error)
	listPostsByAuthorFunc func(ctx context.Context, authorID string) ([]domain.Post, error)
	updatePostFunc        func(ctx context.Context, requesterID string, postID string, content string) (*domain.Post, error)
	deletePostFunc        func(ctx context.Context, requesterID string, postID string) error
}

func (m *mockPostService) CreatePost(ctx context.Context, authorID string, content string) (*domain.Post, error) {
	if m.createPostFunc != nil {
		return m.createPostFunc(ctx, authorID, content)
	}
	return nil, nil
}

func (m *mockPostService) GetPost(ctx context.Context, id string) (*domain.Post, error) {
	if m.getPostFunc != nil {
		return m.getPostFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockPostService) ListPostsByAuthor(ctx context.Context, authorID string) ([]domain.Post, error) {
	if m.listPostsByAuthorFunc != nil {
		return m.listPostsByAuthorFunc(ctx, authorID)
	}
	return nil, nil
}

func (m *mockPostService) UpdatePost(ctx context.Context, requesterID string, postID string, content string) (*domain.Post, error) {
	if m.updatePostFunc != nil {
		return m.updatePostFunc(ctx, requesterID, postID, content)
	}
	return nil, nil
}

func (m *mockPostService) DeletePost(ctx context.Context, requesterID string, postID string) error {
	if m.deletePostFunc != nil {
		return m.deletePostFunc(ctx, requesterID, postID)
	}
	return nil
}

func requestContextWithIDs() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.RequestIDKey, "req-test")
	ctx = context.WithValue(ctx, middleware.AuthenticatedUserIDKey, "user-1")
	return ctx
}

func TestCreatePost_Unauthenticated(t *testing.T) {
	h := NewPostHandler(&mockPostService{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{"content":"hello"}`))
	r = r.WithContext(context.WithValue(context.Background(), middleware.RequestIDKey, "req-test"))

	h.CreatePost(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestCreatePost_Created(t *testing.T) {
	h := NewPostHandler(&mockPostService{
		createPostFunc: func(ctx context.Context, authorID string, content string) (*domain.Post, error) {
			if authorID != "user-1" || content != "hello" {
				t.Fatalf("unexpected create args: authorID=%q content=%q", authorID, content)
			}
			return &domain.Post{ID: "post-1", AuthorID: authorID, Content: content}, nil
		},
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{"content":"hello"}`))
	r = r.WithContext(requestContextWithIDs())

	h.CreatePost(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusCreated, w.Body.String())
	}
	var body apiresponse.SuccessEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if !body.Success {
		t.Fatal("expected success envelope")
	}
}

func TestCreatePost_ValidationError(t *testing.T) {
	h := NewPostHandler(&mockPostService{
		createPostFunc: func(ctx context.Context, authorID string, content string) (*domain.Post, error) {
			return nil, service.ErrValidation
		},
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{"content":""}`))
	r = r.WithContext(requestContextWithIDs())

	h.CreatePost(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	var body apiresponse.ErrorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Error != apperrors.CodeValidationError || body.Status != http.StatusBadRequest {
		t.Fatalf("unexpected error body: %+v", body)
	}
}

func TestGetPost_OK(t *testing.T) {
	h := NewPostHandler(&mockPostService{
		getPostFunc: func(ctx context.Context, id string) (*domain.Post, error) {
			if id != "post-1" {
				t.Fatalf("id = %q, want %q", id, "post-1")
			}
			return &domain.Post{ID: "post-1", AuthorID: "user-1", Content: "hello"}, nil
		},
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/posts/post-1", nil)
	r = r.WithContext(context.WithValue(context.Background(), middleware.RequestIDKey, "req-test"))

	h.GetPost(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestListPostsByAuthor_OK(t *testing.T) {
	h := NewPostHandler(&mockPostService{
		listPostsByAuthorFunc: func(ctx context.Context, authorID string) ([]domain.Post, error) {
			if authorID != "user-1" {
				t.Fatalf("authorID = %q, want %q", authorID, "user-1")
			}
			return []domain.Post{{ID: "post-1"}, {ID: "post-2"}}, nil
		},
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/posts?authorId=user-1", nil)
	r = r.WithContext(context.WithValue(context.Background(), middleware.RequestIDKey, "req-test"))

	h.ListPostsByAuthor(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestUpdatePost_Forbidden(t *testing.T) {
	h := NewPostHandler(&mockPostService{
		updatePostFunc: func(ctx context.Context, requesterID string, postID string, content string) (*domain.Post, error) {
			return nil, service.ErrForbidden
		},
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/posts/post-1", strings.NewReader(`{"content":"edited"}`))
	r = r.WithContext(requestContextWithIDs())

	h.UpdatePost(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusForbidden, w.Body.String())
	}
}

func TestUpdatePost_ValidationError(t *testing.T) {
	h := NewPostHandler(&mockPostService{
		updatePostFunc: func(ctx context.Context, requesterID string, postID string, content string) (*domain.Post, error) {
			return nil, service.ErrValidation
		},
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/posts/post-1", strings.NewReader(`{"content":"   "}`))
	r = r.WithContext(requestContextWithIDs())

	h.UpdatePost(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	var body apiresponse.ErrorEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Error != apperrors.CodeValidationError || body.Status != http.StatusBadRequest {
		t.Fatalf("unexpected error body: %+v", body)
	}
}

func TestDeletePost_NoContent(t *testing.T) {
	h := NewPostHandler(&mockPostService{
		deletePostFunc: func(ctx context.Context, requesterID string, postID string) error {
			if requesterID != "user-1" || postID != "post-1" {
				t.Fatalf("unexpected delete args: requesterID=%q postID=%q", requesterID, postID)
			}
			return nil
		},
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/posts/post-1", nil)
	r = r.WithContext(requestContextWithIDs())

	h.DeletePost(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusNoContent, w.Body.String())
	}
	if w.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", w.Body.String())
	}
}

func TestDeletePost_NotFound(t *testing.T) {
	h := NewPostHandler(&mockPostService{
		deletePostFunc: func(ctx context.Context, requesterID string, postID string) error {
			return service.ErrPostNotFound
		},
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/posts/post-1", nil)
	r = r.WithContext(requestContextWithIDs())

	h.DeletePost(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d body=%s", w.Code, http.StatusNotFound, w.Body.String())
	}
}
