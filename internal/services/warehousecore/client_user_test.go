package warehousecore

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUser_ListUsers(t *testing.T) {
	user := User{ID: 5, Username: "alice"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/admin/users/5":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user) //nolint:errcheck
			return
		case "/admin/users":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]User{user}) //nolint:errcheck
			return
		default:
			http.NotFound(w, r)
			return
		}
	}))
	defer srv.Close()

	c := NewClientWithConfig(srv.URL, "")

	got, err := c.GetUser(5)
	if err != nil {
		t.Fatalf("GetUser() unexpected error: %v", err)
	}
	if got.ID != user.ID || got.Username != user.Username {
		t.Fatalf("GetUser() = %+v, want %+v", got, user)
	}

	list, err := c.ListUsers("")
	if err != nil {
		t.Fatalf("ListUsers() unexpected error: %v", err)
	}
	if len(list) != 1 || list[0].ID != user.ID {
		t.Fatalf("ListUsers() = %+v, want [%+v]", list, user)
	}
}
