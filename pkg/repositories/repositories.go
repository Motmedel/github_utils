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
	branch string,
	token string,
	client motmedelHttpUtils.HttpClient,
) ([]byte, *motmedelHttpTypes.HttpContext, error) {
	if owner == "" {
		return nil, nil, githubUtilsErrors.ErrEmptyOwner
	}

	if name == "" {
		return nil, nil, githubUtilsErrors.ErrEmptyName
	}

	if branch == "" {
		return nil, nil, githubUtilsErrors.ErrEmptyBranch
	}

	if token == "" {
		return nil, nil, githubUtilsErrors.ErrEmptyToken
	}

	if client == nil {
		return nil, nil, motmedelHttpErrors.ErrNilHttpClient
	}

	if reposBaseUrl == nil {
		return nil, nil, githubUtilsErrors.ErrNilReposBaseUrl
	}

	requestMethod := http.MethodGet

	requestUrl := *reposBaseUrl
	// TODO: I could do additional validation here.
	ownerSegment := url.PathEscape(owner)
	nameSegment := url.PathEscape(name)
	branchSegment := url.PathEscape(branch)
	requestUrl.Path = path.Join(requestUrl.Path, ownerSegment, nameSegment, "tarball", branchSegment)

	requestUrlString := requestUrl.String()

	httpContext, err := motmedelHttpUtils.SendRequest(
		client,
		requestMethod,
		requestUrlString,
		nil,
		func(request *http.Request) error {
			if request == nil {
				return motmedelHttpErrors.ErrNilHttpRequest
			}

			requestHeader := request.Header
			if requestHeader == nil {
				return motmedelHttpErrors.ErrNilHttpRequestHeader
			}

			requestHeader.Set("Accept", "application/vnd.github+json")
			requestHeader.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			requestHeader.Set("X-GitHub-Api-Version", "2022-11-28")

			return nil
		},
	)
	if err != nil {
		return nil, httpContext, &motmedelErrors.InputError{
			Message: "An error occurred when sending the request.",
			Cause:   err,
			Input:   []any{requestMethod, requestUrlString},
		}
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
		return nil, httpContext, &motmedelErrors.InputError{
			Message: "An error occurred when getting the tarball.",
			Cause:   err,
			Input:   []any{owner, name, branch, client},
		}
	}
	if len(tarball) == 0 {
		return nil, nil, nil
	}

	response := httpContext.Response
	if response == nil {
		return nil, httpContext, motmedelHttpErrors.ErrNilHttpResponse
	}

	responseHeader := response.Header
	if responseHeader == nil {
		return nil, httpContext, motmedelHttpErrors.ErrNilHttpResponseHeader
	}

	if contentTypeValue := responseHeader.Get("Content-Type"); contentTypeValue != expectedTarballContentType {
		return nil, httpContext, &motmedelErrors.InputError{
			Message: "The HTTP response does not have the expected content type.",
			Input:   contentTypeValue,
		}
	}

	tarballGzipReader, err := gzip.NewReader(bytes.NewReader(tarball))
	if err != nil {
		return nil, httpContext, &motmedelErrors.CauseError{
			Message: "An error occurred when creating the tarball gzip reader.",
			Cause:   err,
		}
	}

	return tarballGzipReader, httpContext, nil
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
		return nil, httpContext, &motmedelErrors.InputError{
			Message: "An error occurred when obtaining the tarball reader.",
			Cause:   err,
			Input:   []any{owner, repository, branch},
		}
	}
	if tarballReader == nil {
		return nil, httpContext, githubUtilsErrors.ErrNilTarballReader
	}

	archive, err := motmedelTar.MakeArchiveFromReader(tarballReader)
	if err != nil {
		return nil, httpContext, &motmedelErrors.InputError{
			Message: "An occurred when making an archive from the tarball reader.",
			Cause:   err,
			Input:   tarballReader,
		}
	}

	return archive, httpContext, nil
}

func getTarballPrefix(response *http.Response) (string, error) {
	if response == nil {
		return "", nil
	}

	responseHeader := response.Header
	if responseHeader == nil {
		return "", motmedelHttpErrors.ErrNilHttpResponseHeader
	}

	contentDispositionValue := responseHeader.Get("Content-Disposition")
	contentDispositionBytes := []byte(contentDispositionValue)
	contentDisposition, err := content_disposition.ParseContentDisposition(contentDispositionBytes)
	if err != nil {
		return "", &motmedelErrors.InputError{
			Message: "An error occurred when obtaining a content disposition.",
			Cause:   err,
			Input:   contentDispositionBytes,
		}
	}
	if contentDisposition == nil {
		return "", githubUtilsErrors.ErrNilContentDisposition
	}

	contentDispositionFilename := contentDisposition.Filename
	if contentDispositionFilename == "" {
		return "", githubUtilsErrors.ErrEmptyContentDispositionFilename
	}

	tarballPrefix, _, _ := strings.Cut(contentDispositionFilename, ".")

	return tarballPrefix, nil
}

func GetUnprefixedTarArchive(
	owner string,
	repository string,
	branch string,
	token string,
	httpClient motmedelHttpUtils.HttpClient,
) (motmedelTarTypes.Archive, *motmedelHttpTypes.HttpContext, error) {

	tarArchive, httpContext, err := GetTarArchive(owner, repository, branch, token, httpClient)
	if err != nil {
		return nil, httpContext, &motmedelErrors.InputError{
			Message: "An error occurred when obtaining the repository tar archive.",
			Cause:   err,
			Input:   []any{owner, repository, branch},
		}
	}
	if len(tarArchive) == 0 {
		return nil, httpContext, nil
	}

	response := httpContext.Response
	repositoryTarballPrefix, err := getTarballPrefix(response)
	if err != nil {
		return nil, httpContext, &motmedelErrors.InputError{
			Message: "An error occurred when obtaining the GitHub repository tarball prefix.",
			Cause:   err,
			Input:   response,
		}
	}

	unprefixedTarArchive, ok := tarArchive.SetDirectory(repositoryTarballPrefix)
	if !ok {
		return nil, httpContext, motmedelTarErrors.ErrSetDirectoryError
	}

	return unprefixedTarArchive, httpContext, nil
}

func init() {
	var err error
	if reposBaseUrl, err = url.Parse(reposBaseUrlString); err != nil {
		panic(fmt.Sprintf("An error occurred when parsing the repos base url: %v", err))
	}
}
