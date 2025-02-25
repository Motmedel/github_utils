package repositories

import (
	"bytes"
	"compress/gzip"
	"fmt"
	githubUtilsErrors "github.com/Motmedel/github_utils/pkg/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/content_disposition"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
	motmedelTar "github.com/Motmedel/utils_go/pkg/tar"
	motmedelTarErrors "github.com/Motmedel/utils_go/pkg/tar/errors"
	motmedelTarTypes "github.com/Motmedel/utils_go/pkg/tar/types"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const expectedTarballContentType = "application/x-gzip"

const reposBaseUrlString = "https://api.github.com/repos/"

var reposBaseUrl *url.URL

func GetTarball(
	owner string,
	name string,
	ref string,
	token string,
	client motmedelHttpUtils.HttpClient,
) ([]byte, *motmedelHttpTypes.HttpContext, error) {
	if owner == "" {
		return nil, nil, motmedelErrors.MakeErrorWithStackTrace(githubUtilsErrors.ErrEmptyOwner)
	}

	if name == "" {
		return nil, nil, motmedelErrors.MakeErrorWithStackTrace(githubUtilsErrors.ErrEmptyName)
	}

	if ref == "" {
		return nil, nil, motmedelErrors.MakeErrorWithStackTrace(githubUtilsErrors.ErrEmptyRef)
	}

	if token == "" {
		return nil, nil, motmedelErrors.MakeErrorWithStackTrace(githubUtilsErrors.ErrEmptyToken)
	}

	if client == nil {
		return nil, nil, motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpClient)
	}

	if reposBaseUrl == nil {
		return nil, nil, motmedelErrors.MakeErrorWithStackTrace(githubUtilsErrors.ErrNilReposBaseUrl)
	}

	requestMethod := http.MethodGet

	requestUrl := *reposBaseUrl
	// TODO: I could do additional validation here.
	ownerSegment := url.PathEscape(owner)
	nameSegment := url.PathEscape(name)
	branchSegment := url.PathEscape(ref)
	requestUrl.Path = path.Join(requestUrl.Path, ownerSegment, nameSegment, "tarball", branchSegment)

	requestUrlString := requestUrl.String()

	httpContext, err := motmedelHttpUtils.SendRequest(
		client,
		requestMethod,
		requestUrlString,
		nil,
		func(request *http.Request) error {
			// TODO: Use proper errors.
			if request == nil {
				return motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpRequest)
			}

			requestHeader := request.Header
			if requestHeader == nil {
				return motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpRequestHeader)
			}

			requestHeader.Set("Accept", "application/vnd.github+json")
			requestHeader.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			requestHeader.Set("X-GitHub-Api-Version", "2022-11-28")

			return nil
		},
	)
	if err != nil {
		return nil, httpContext, motmedelErrors.MakeError(
			fmt.Errorf("send request: %w", err),
			requestMethod, requestUrlString,
		)
	}

	if httpContext == nil {
		return nil, nil, nil
	}

	return httpContext.ResponseBody, httpContext, nil
}

func GetTarballReader(
	owner string,
	name string,
	branch string,
	token string,
	client motmedelHttpUtils.HttpClient,
) (*gzip.Reader, *motmedelHttpTypes.HttpContext, error) {
	tarball, httpContext, err := GetTarball(owner, name, branch, token, client)
	if err != nil {
		return nil, httpContext, motmedelErrors.MakeError(
			fmt.Errorf("get tarball: %w", err),
			owner, name, branch,
		)
	}
	if len(tarball) == 0 {
		return nil, nil, nil
	}

	response := httpContext.Response
	if response == nil {
		return nil, httpContext, motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpResponse)
	}

	responseHeader := response.Header
	if responseHeader == nil {
		return nil, httpContext, motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpResponseHeader)
	}

	if contentTypeValue := responseHeader.Get("Content-Type"); contentTypeValue != expectedTarballContentType {
		return nil, httpContext, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf("%w: %s", githubUtilsErrors.ErrUnexpectedContentType, contentTypeValue),
			contentTypeValue,
		)
	}

	tarballReader, err := gzip.NewReader(bytes.NewReader(tarball))
	if err != nil {
		return nil, httpContext, motmedelErrors.MakeErrorWithStackTrace(fmt.Errorf("gzip new reader: %w", err))
	}

	return tarballReader, httpContext, nil
}

func GetTarArchive(
	owner string,
	repository string,
	branch string,
	token string,
	httpClient motmedelHttpUtils.HttpClient,
) (motmedelTarTypes.Archive, *motmedelHttpTypes.HttpContext, error) {
	tarballReader, httpContext, err := GetTarballReader(owner, repository, branch, token, httpClient)
	if err != nil {
		return nil, httpContext, motmedelErrors.MakeError(
			fmt.Errorf("get tarball reader: %w", err),
			owner, repository, branch,
		)
	}
	if tarballReader == nil {
		return nil, httpContext, githubUtilsErrors.ErrNilTarballReader
	}

	archive, err := motmedelTar.MakeArchiveFromReader(tarballReader)
	if err != nil {
		return nil, httpContext, motmedelErrors.MakeError(
			fmt.Errorf("make archive from reader: %w", err),
			tarballReader,
		)
	}

	return archive, httpContext, nil
}

func getTarballPrefix(response *http.Response) (string, error) {
	if response == nil {
		return "", nil
	}

	responseHeader := response.Header
	if responseHeader == nil {
		return "", motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpResponseHeader)
	}

	contentDispositionValue := responseHeader.Get("Content-Disposition")
	contentDispositionBytes := []byte(contentDispositionValue)
	contentDisposition, err := content_disposition.ParseContentDisposition(contentDispositionBytes)
	if err != nil {
		return "", motmedelErrors.MakeError(
			fmt.Errorf("parse content disposition: %w", err),
			contentDispositionBytes,
		)
	}
	if contentDisposition == nil {
		return "", motmedelErrors.MakeErrorWithStackTrace(githubUtilsErrors.ErrNilContentDisposition)
	}

	contentDispositionFilename := contentDisposition.Filename
	if contentDispositionFilename == "" {
		return "", motmedelErrors.MakeErrorWithStackTrace(githubUtilsErrors.ErrEmptyContentDispositionFilename)
	}

	tarballPrefix, _, _ := strings.Cut(contentDispositionFilename, ".")

	return tarballPrefix, nil
}

func GetUnprefixedTarArchive(
	owner string,
	name string,
	branch string,
	token string,
	httpClient motmedelHttpUtils.HttpClient,
) (motmedelTarTypes.Archive, *motmedelHttpTypes.HttpContext, error) {

	archive, httpContext, err := GetTarArchive(owner, name, branch, token, httpClient)
	if err != nil {
		return nil, httpContext, motmedelErrors.MakeError(
			fmt.Errorf("get tar archive: %w", err),
			owner, name, branch,
		)
	}
	if len(archive) == 0 {
		return nil, httpContext, nil
	}

	response := httpContext.Response
	prefix, err := getTarballPrefix(response)
	if err != nil {
		return nil, httpContext, motmedelErrors.MakeError(
			fmt.Errorf("get tarball prefix: %w", err),
			response,
		)
	}
	if prefix == "" {
		return nil, httpContext, motmedelErrors.MakeErrorWithStackTrace(githubUtilsErrors.ErrEmptyTarballPrefix)
	}

	unprefixedArchive, ok := archive.SetDirectory(prefix)
	if !ok {
		return nil, httpContext, motmedelErrors.MakeErrorWithStackTrace(motmedelTarErrors.ErrSetDirectoryError)
	}

	return unprefixedArchive, httpContext, nil
}

func init() {
	var err error
	if reposBaseUrl, err = url.Parse(reposBaseUrlString); err != nil {
		panic(fmt.Sprintf("An error occurred when parsing the repos base url: %v", err))
	}
}
