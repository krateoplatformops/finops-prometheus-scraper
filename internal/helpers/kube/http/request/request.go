package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/http/request"
	"github.com/krateoplatformops/plumbing/http/util"
	"github.com/krateoplatformops/plumbing/ptr"

	localstatus "github.com/krateoplatformops/finops-prometheus-scraper/internal/helpers/kube/http/response"
)

const maxUnstructuredResponseTextBytes = 2048

func Do(ctx context.Context, opts request.RequestOptions) *localstatus.Status {
	uri := strings.TrimSuffix(opts.Endpoint.ServerURL, "/")
	if len(opts.Path) > 0 {
		uri = fmt.Sprintf("%s/%s", uri, strings.TrimPrefix(opts.Path, "/"))
	}

	u, err := url.Parse(uri)
	if err != nil {
		return localstatus.New(http.StatusInternalServerError, nil, err)
	}

	verb := ptr.Deref(opts.Verb, http.MethodGet)

	var body io.Reader
	if s := ptr.Deref(opts.Payload, ""); len(s) > 0 {
		body = strings.NewReader(s)
	}

	call, err := http.NewRequestWithContext(ctx, verb, u.String(), body)
	if err != nil {
		return localstatus.New(http.StatusInternalServerError, nil, err)
	}
	// Additional headers for AWS Signature 4 algorithm
	if opts.Endpoint.HasAwsAuth() {
		headers, _, _, _, _, _ := request.ComputeAwsHeaders(opts.Endpoint, &opts.RequestInfo)
		opts.Headers = append(opts.Headers, headers...)
		opts.Headers = append(opts.Headers, xcontext.LabelKrateoTraceId+":"+xcontext.TraceId(ctx, true))
		for i := range opts.Headers {
			hParts := strings.Split(opts.Headers[i], ":")
			opts.Headers[i] = strings.ToLower(strings.Trim(hParts[0], " ")) + ":" + strings.Trim(hParts[1], " ")
		}
		sort.Strings(opts.Headers)
	} else {
		call.Header.Set(xcontext.LabelKrateoTraceId, xcontext.TraceId(ctx, true))
	}

	// log.Info().Msgf("Header: %s", opts.Headers)
	if len(opts.Headers) > 0 {
		for _, el := range opts.Headers {
			idx := strings.Index(el, ":")
			if idx <= 0 {
				continue
			}
			key := el[:idx]
			val := strings.TrimSpace(el[idx+1:])
			call.Header.Set(key, val)
		}
	}

	cli, err := request.HTTPClientForEndpoint(opts.Endpoint, &opts.RequestInfo)
	if err != nil {
		return localstatus.New(http.StatusInternalServerError, nil,
			fmt.Errorf("unable to create HTTP Client for endpoint: %w", err))
	}

	// Wrap the existing client in a RetryClient
	retryCli := util.NewRetryClient(cli)

	// Use RetryClient instead of the raw client  cli.Do(call)
	respo, err := retryCli.Do(call)
	if err != nil {
		return localstatus.New(http.StatusInternalServerError, nil, err)
	}
	defer respo.Body.Close()

	statusOK := respo.StatusCode >= 200 && respo.StatusCode < 300
	if !statusOK {
		dat, err := io.ReadAll(io.LimitReader(respo.Body, maxUnstructuredResponseTextBytes))
		if err != nil {
			return localstatus.New(http.StatusInternalServerError, nil, err)
		}
		// log.Info().Msgf("Body plumbing: %s", string(dat))

		res := &localstatus.Status{}
		if err := json.Unmarshal(dat, res); err != nil {
			res = localstatus.New(respo.StatusCode, &respo.Header, fmt.Errorf("%s", string(dat)))
			return res
		}

		return res
	}

	// Service providers REST APIs do not always follow the standard JSON response, hence these lines are commented
	// if ct := respo.Header.Get("Content-Type"); !strings.Contains(ct, "json") {
	// 	return response.New(http.StatusNotAcceptable, fmt.Errorf("content type %q is not allowed", ct))
	// }

	if opts.ResponseHandler != nil {
		if err := opts.ResponseHandler(respo.Body); err != nil {
			return localstatus.New(http.StatusInternalServerError, nil, err)
		}
		return localstatus.New(http.StatusOK, &respo.Header, nil)
	}

	return localstatus.New(http.StatusNoContent, &respo.Header, nil)
}
