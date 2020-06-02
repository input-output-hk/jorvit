package webproxy

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/input-output-hk/jorvit/internal/datastore"
)

var (
	reverseProxyAddress = "http://127.0.0.1:8001"
	proposals           datastore.ProposalsStore
	funds               datastore.FundsStore
	block0Bin           *[]byte
)

// ShiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
func ShiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}

type App struct {
	// Not using http.Handler for decoupling
	ApiHandler *ApiHandler
}

func (h *App) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = ShiftPath(req.URL.Path)

	switch head {
	case "api":
		h.ApiHandler.ServeHTTP(res, req)
		return
	default:
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}
}

type ApiHandler struct {
	V0Handler *V0Handler
}

func (h *ApiHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = ShiftPath(req.URL.Path)

	switch head {
	case "v0":
		h.V0Handler.ServeHTTP(res, req)
		return
	default:
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}
}

type V0Handler struct {
	ProposalHandler *ProposalHandler
	Block0Handler   *Block0Handler
	FundInfoHandler *FundInfoHandler
}

func (h *V0Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = ShiftPath(req.URL.Path)

	res.Header().Set("Access-Control-Allow-Origin", "*")

	switch head {
	case "proposals":
		h.ProposalHandler.ServeHTTP(res, req)
		return
	case "block0":
		h.Block0Handler.ServeHTTP(res, req)
		return
	case "fund":
		h.FundInfoHandler.ServeHTTP(res, req)
	case "account":
		serveReverseProxy("/api/v0/account", res, req)
		return
	case "block":
		serveReverseProxy("/api/v0/block", res, req)
		return
	case "fragment":
		serveReverseProxy("/api/v0/fragment", res, req)
		return
	case "message":
		serveReverseProxy("/api/v0/message", res, req)
		return
	case "settings":
		serveReverseProxy("/api/v0/settings", res, req)
		return
	default:
		http.Error(res, "Not Found", http.StatusNotFound)
		return
	}
}

type ProposalHandler struct {
	ProposalListAll    *ProposalListAll
	ProposalListSingle *ProposalListSingle
}

func (h *ProposalHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head, internalID string
	head, req.URL.Path = ShiftPath(req.URL.Path)
	internalID = head

	if req.URL.Path == "/" {
		switch internalID {
		case "":
			h.ProposalListAll.ServeHTTP(res, req)
			return
		default:
			h.ProposalListSingle.Handler(internalID, res, req).ServeHTTP(res, req)
			return
		}
	} else /* if req.URL.Path != "/" */ {
		head, tail := ShiftPath(req.URL.Path)
		_ = tail
		switch head {
		default:
			http.Error(res, "Not Found", http.StatusNotFound)
			return
		}
	}
}

type ProposalListAll struct {
}

func (h *ProposalListAll) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	_ = req.URL.Query()

	switch req.Method {
	case "GET":
		if proposals.Total() == 0 {
			res.WriteHeader(http.StatusNotFound)
			res.Write([]byte(`{"error": "empty data"}`))
			return
		}
		resData, err := json.MarshalIndent(proposals.All(), "", "  ")
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(`{"error": "error marshalling data"}`))
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write(resData)
		return
	default:
		http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
	}
}

type ProposalListSingle struct {
}

func (h *ProposalListSingle) Handler(internalID string, res http.ResponseWriter, req *http.Request) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		_ = req.URL.Query()

		switch req.Method {
		case "GET":
			proposal := proposals.SearchID(internalID)
			if proposal == nil {
				res.WriteHeader(http.StatusNotFound)
				res.Write([]byte(`{"error": not found"}`))
				return
			}
			resData, err := json.MarshalIndent(proposal, "", "  ")
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte(`{"error": "error marshalling data"}`))
				return
			}
			res.WriteHeader(http.StatusOK)
			res.Write(resData)
			return
		default:
			http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
		}
	})
}

type Block0Handler struct {
}

func (h *Block0Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/octet-stream")
	res.Header().Set("Content-Length", strconv.Itoa(len(*block0Bin)))
	switch req.Method {
	case "GET":
		res.WriteHeader(http.StatusOK)
		res.Write(*block0Bin)
		return
	default:
		http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
	}
}

type FundInfoHandler struct {
}

func (h *FundInfoHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "GET":
		if funds.Total() == 0 {
			res.WriteHeader(http.StatusNotFound)
			res.Write([]byte(`{"error": "empty data"}`))
			return
		}
		resData, err := json.MarshalIndent(funds.First(), "", "  ")
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(`{"error": "error marshalling data"}`))
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write(resData)
		return
	default:
		http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
	}
}

func Run(p datastore.ProposalsStore, f datastore.FundsStore, block0 *[]byte, address string, revProxyAddr string) error {
	proposals = p
	funds = f
	reverseProxyAddress = revProxyAddr
	block0Bin = block0

	app := &App{
		ApiHandler: &ApiHandler{
			V0Handler: &V0Handler{
				ProposalHandler: &ProposalHandler{
					ProposalListAll:    new(ProposalListAll),
					ProposalListSingle: new(ProposalListSingle),
				},
				Block0Handler:   new(Block0Handler),
				FundInfoHandler: new(FundInfoHandler),
			},
		},
	}

	srv := &http.Server{
		Addr:    address,
		Handler: app,
	}

	return srv.ListenAndServe()
}

// Serve a reverse proxy for a given url
func serveReverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(reverseProxyAddress + target)

	proxy := httputil.NewSingleHostReverseProxy(url)

	// SSL redirection
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	proxy.ServeHTTP(res, req)
}
