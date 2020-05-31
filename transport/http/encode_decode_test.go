package http

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/moogar0880/problems"
)

func TestNopResponseEncoder(t *testing.T) {
	handler := kithttp.NewServer(
		func(context.Context, interface{}) (interface{}, error) { return "response", nil },
		kithttp.NopRequestDecoder,
		NopResponseEncoder,
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("unexpected status code\nexpected: %d\nactual:   %d", want, have)
	}
}

func TestStatusCodeResponseEncoder(t *testing.T) {
	handler := kithttp.NewServer(
		func(context.Context, interface{}) (interface{}, error) { return "response", nil },
		kithttp.NopRequestDecoder,
		StatusCodeResponseEncoder(http.StatusNoContent),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if want, have := http.StatusNoContent, resp.StatusCode; want != have {
		t.Errorf("unexpected status code\nexpected: %d\nactual:   %d", want, have)
	}
}

func TestJSONResponseEncoder(t *testing.T) {
	handler := kithttp.NewServer(
		func(context.Context, interface{}) (interface{}, error) {
			return struct {
				Foo string `json:"foo"`
			}{Foo: "bar"}, nil
		},
		kithttp.NopRequestDecoder,
		JSONResponseEncoder,
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if want, have := http.StatusOK, resp.StatusCode; want != have {
		t.Errorf("unexpected status code\nexpected: %d\nactual:   %d", want, have)
	}

	buf, _ := ioutil.ReadAll(resp.Body)
	if want, have := `{"foo":"bar"}`, strings.TrimSpace(string(buf)); want != have {
		t.Errorf("unexpected body\nexpected: %s\nactual:   %s", want, have)
	}
}

func TestWithStatusCode(t *testing.T) {
	type resp struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	}

	response := WithStatusCode(resp{"id", "text"}, http.StatusCreated)

	statusCoder, ok := response.(kithttp.StatusCoder)
	if !ok {
		t.Fatal("response was expected to be a StatusCoder")
	}

	if want, have := http.StatusCreated, statusCoder.StatusCode(); want != have {
		t.Errorf("unexpected status code\nactual:   %d\nexpected: %d", have, want)
	}

	body, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}

	expectedBody := `{"id":"id","text":"text"}`
	if want, have := expectedBody, string(body); want != have {
		t.Errorf("unexpected body\nexpected: %s\nactual:   %s", want, have)
	}
}

type failer struct {
	err error
}

func (f failer) Failed() error {
	return f.err
}

func TestErrorResponseEncoder(t *testing.T) {
	t.Parallel()

	t.Run("response", func(t *testing.T) {
		handler := kithttp.NewServer(
			func(context.Context, interface{}) (interface{}, error) {
				return struct {
					Foo string `json:"foo"`
				}{Foo: "bar"}, nil
			},
			kithttp.NopRequestDecoder,
			ErrorResponseEncoder(JSONResponseEncoder, func(i context.Context, w http.ResponseWriter, e error) error {
				problem := problems.NewDetailedProblem(http.StatusBadRequest, e.Error())

				w.Header().Set("Content-Type", problems.ProblemMediaType)
				w.WriteHeader(problem.Status)

				return json.NewEncoder(w).Encode(problem)
			}),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if want, have := http.StatusOK, resp.StatusCode; want != have {
			t.Errorf("unexpected status code\nexpected: %d\nactual:   %d", want, have)
		}

		buf, _ := ioutil.ReadAll(resp.Body)
		if want, have := `{"foo":"bar"}`, strings.TrimSpace(string(buf)); want != have {
			t.Errorf("unexpected body\nexpected: %s\nactual:   %s", want, have)
		}
	})

	t.Run("error", func(t *testing.T) {
		handler := kithttp.NewServer(
			func(context.Context, interface{}) (interface{}, error) {
				return failer{errors.New("error")}, nil
			},
			kithttp.NopRequestDecoder,
			ErrorResponseEncoder(JSONResponseEncoder, func(i context.Context, w http.ResponseWriter, e error) error {
				problem := problems.NewDetailedProblem(http.StatusBadRequest, e.Error())

				w.Header().Set("Content-Type", problems.ProblemMediaType)
				w.WriteHeader(problem.Status)

				return json.NewEncoder(w).Encode(problem)
			}),
		)

		server := httptest.NewServer(handler)
		defer server.Close()

		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if want, have := http.StatusBadRequest, resp.StatusCode; want != have {
			t.Errorf("unexpected status code\nexpected: %d\nactual:   %d", want, have)
		}

		expectedBody := `{"type":"about:blank","title":"Bad Request","status":400,"detail":"error"}`
		buf, _ := ioutil.ReadAll(resp.Body)
		if want, have := expectedBody, strings.TrimSpace(string(buf)); want != have {
			t.Errorf("unexpected body\nexpected: %s\nactual:   %s", want, have)
		}
	})
}

type problemConverterStub struct {
	problem problems.Problem
}

func (s problemConverterStub) NewProblem(_ context.Context, _ error) interface{} {
	return s.problem
}

func testStatusAndContentType(t *testing.T, resp *http.Response, status int, contentType string) {
	t.Helper()

	if want, have := status, resp.StatusCode; want != have {
		t.Errorf("unexpected status\nexpected: %d\nactual:   %d", want, have)
	}

	if want, have := contentType, resp.Header.Get("Content-Type"); want != have {
		t.Errorf("unexpected content type\nexpected: %s\nactual:   %s", want, have)
	}
}

// nolint: dupl
func TestNewJSONProblemErrorEncoder(t *testing.T) {
	t.Run("without_status", func(t *testing.T) {
		problemConverter := problemConverterStub{problems.NewProblem()}

		errorEncoder := NewJSONProblemErrorEncoder(problemConverter)

		w := httptest.NewRecorder()

		errorEncoder(context.Background(), errors.New("error"), w)

		resp := w.Result()
		defer resp.Body.Close()

		testStatusAndContentType(t, resp, http.StatusOK, problems.ProblemMediaType)
	})

	t.Run("with_empty_status", func(t *testing.T) {
		problemConverter := problemConverterStub{problems.NewDetailedProblem(0, "error")}

		errorEncoder := NewJSONProblemErrorEncoder(problemConverter)

		w := httptest.NewRecorder()

		errorEncoder(context.Background(), errors.New("error"), w)

		resp := w.Result()
		defer resp.Body.Close()

		testStatusAndContentType(t, resp, http.StatusOK, problems.ProblemMediaType)

		var details struct {
			Detail string `json:"detail"`
		}

		err := json.NewDecoder(resp.Body).Decode(&details)
		if err != nil {
			t.Fatal(err)
		}

		if want, have := "error", details.Detail; want != have {
			t.Errorf("unexpected detail\nexpected: %s\nactual:   %s", want, have)
		}
	})

	t.Run("with_status", func(t *testing.T) {
		problemConverter := problemConverterStub{problems.NewDetailedProblem(http.StatusNotFound, "error")}

		errorEncoder := NewJSONProblemErrorEncoder(problemConverter)

		w := httptest.NewRecorder()

		errorEncoder(context.Background(), errors.New("error"), w)

		resp := w.Result()
		defer resp.Body.Close()

		testStatusAndContentType(t, resp, http.StatusNotFound, problems.ProblemMediaType)
	})
}

func TestNewDefaultJSONProblemErrorEncoder(t *testing.T) {
	errorEncoder := NewDefaultJSONProblemErrorEncoder()

	w := httptest.NewRecorder()

	errorEncoder(context.Background(), errors.New("error"), w)

	resp := w.Result()
	defer resp.Body.Close()

	testStatusAndContentType(t, resp, http.StatusInternalServerError, problems.ProblemMediaType)

	var details struct {
		Detail string `json:"detail"`
	}

	err := json.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		t.Fatal(err)
	}

	if want, have := "something went wrong", details.Detail; want != have {
		t.Errorf("unexpected detail\nexpected: %s\nactual:   %s", want, have)
	}
}

// nolint: dupl
func TestNewXMLProblemErrorEncoder(t *testing.T) {
	t.Run("without_status", func(t *testing.T) {
		problemConverter := problemConverterStub{problems.NewProblem()}

		errorEncoder := NewXMLProblemErrorEncoder(problemConverter)

		w := httptest.NewRecorder()

		errorEncoder(context.Background(), errors.New("error"), w)

		resp := w.Result()
		defer resp.Body.Close()

		testStatusAndContentType(t, resp, http.StatusOK, problems.ProblemMediaTypeXML)
	})

	t.Run("with_empty_status", func(t *testing.T) {
		problemConverter := problemConverterStub{problems.NewDetailedProblem(0, "error")}

		errorEncoder := NewXMLProblemErrorEncoder(problemConverter)

		w := httptest.NewRecorder()

		errorEncoder(context.Background(), errors.New("error"), w)

		resp := w.Result()
		defer resp.Body.Close()

		testStatusAndContentType(t, resp, http.StatusOK, problems.ProblemMediaTypeXML)

		var details struct {
			Detail string `xml:""`
		}

		err := xml.NewDecoder(resp.Body).Decode(&details)
		if err != nil {
			t.Fatal(err)
		}

		if want, have := "error", details.Detail; want != have {
			t.Errorf("unexpected detail\nexpected: %s\nactual:   %s", want, have)
		}
	})

	t.Run("with_status", func(t *testing.T) {
		problemConverter := problemConverterStub{problems.NewDetailedProblem(http.StatusNotFound, "error")}

		errorEncoder := NewXMLProblemErrorEncoder(problemConverter)

		w := httptest.NewRecorder()

		errorEncoder(context.Background(), errors.New("error"), w)

		resp := w.Result()
		defer resp.Body.Close()

		testStatusAndContentType(t, resp, http.StatusNotFound, problems.ProblemMediaTypeXML)
	})
}

func TestNewDefaultXMLProblemErrorEncoder(t *testing.T) {
	errorEncoder := NewDefaultXMLProblemErrorEncoder()

	w := httptest.NewRecorder()

	errorEncoder(context.Background(), errors.New("error"), w)

	resp := w.Result()
	defer resp.Body.Close()

	testStatusAndContentType(t, resp, http.StatusInternalServerError, problems.ProblemMediaTypeXML)

	var details struct {
		Detail string `xml:""`
	}

	err := xml.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		t.Fatal(err)
	}

	if want, have := "something went wrong", details.Detail; want != have {
		t.Errorf("unexpected detail\nexpected: %s\nactual:   %s", want, have)
	}
}
