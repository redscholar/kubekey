package apiserver

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/cmd/apiserver/app/options"
	serverinfrastructurev1beta1 "github.com/kubesphere/kubekey/v4/pkg/apiserver/infrastructure/v1beta1"
	serverkkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apiserver/kkcore/v1"
	serverresourcesv1 "github.com/kubesphere/kubekey/v4/pkg/apiserver/resources/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

func NewAPIServer(o *options.APIServerOptions) *APIServer {
	return &APIServer{
		Port:       o.Port,
		Workdir:    o.Workdir,
		SchemaPath: o.SchemaPath,
		UIPath:     o.UIPath,
	}
}

type APIServer struct {
	Port       int
	Workdir    string
	SchemaPath string
	UIPath     string

	ctrlclient.Client
	*rest.Config

	container *restful.Container
	server    *http.Server
}

func (s *APIServer) prepareRun() error {
	s.container = restful.DefaultContainer
	restful.RegisterEntityAccessor(string(types.MergePatchType), restful.NewEntityAccessorJSON(restful.MIME_JSON))
	restful.RegisterEntityAccessor(string(types.JSONPatchType), restful.NewEntityAccessorJSON(restful.MIME_JSON))
	// Initialize REST config for Kubernetes client
	restconfig, err := ctrl.GetConfig()
	if err := proxy.RestConfig(filepath.Join(s.Workdir, _const.RuntimeDir), restconfig); err != nil {
		return err
	}

	// Create Kubernetes client with the REST config
	client, err := ctrlclient.New(restconfig, ctrlclient.Options{
		Scheme: _const.Scheme,
	})
	if err != nil {
		return err
	}

	s.Client = client
	s.Config = restconfig
	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.Port),
		Handler:           s.container,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attack by timing out slow headers
	}

	s.installWebInstallerApi()

	s.server.Handler = s.container
	return nil
}

func (s *APIServer) installWebInstallerApi() {
	s.container.Filter(logRequestAndResponse)
	s.container.RecoverHandler(func(panicReason any, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})
	serverinfrastructurev1beta1.AddToContainer(s.container, s.Client, s.Config, s.Workdir)
	serverresourcesv1.AddToContainer(s.container, s.Client, s.Config, s.Workdir, s.SchemaPath)
	serverkkcorev1.AddToContainer(s.container, s.Client, s.Config, s.Workdir)
}

// Run starts the web server and handles incoming requests
func (s *APIServer) Run(ctx context.Context) error {
	s.prepareRun()

	shutdownCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ctx.Done()
		_ = s.server.Shutdown(shutdownCtx)
	}()

	return s.server.ListenAndServe()
}

// logStackOnRecover handles panic recovery and logs the stack trace
func logStackOnRecover(panicReason any, w http.ResponseWriter) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("recover from panic: %v\n", panicReason))
	for i := 2; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		buf.WriteString(fmt.Sprintf("    %s:%d\n", file, line))
	}
	klog.Errorln(buf.String())

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte("Internal Server Error"))
}

// logRequestAndResponse logs HTTP request and response details
func logRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	// Always log error response
	logWithVerbose := klog.V(4)
	if resp.StatusCode() > http.StatusBadRequest {
		logWithVerbose = klog.V(0)
	}

	logWithVerbose.Infof("%s - \"%s %s %s\" %d %d %dms",
		remoteIP(req.Request),
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)
}

// remoteIP extracts the client IP address from the request, handling various proxy headers
func remoteIP(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get("X-Client-Ip"); ip != "" {
		remoteAddr = ip
	} else if ip := req.Header.Get("X-Real-IP"); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get("X-Forwarded-For"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}
