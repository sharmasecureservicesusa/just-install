package justinstall

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ungerik/go-dry"
)

var expectedContentTypes = []string{
	"application/octet-stream",
	"application/unknown", // Bintray
	"application/x-dosexec",
	"application/x-msdos-program",
	"application/x-msdownload",
	"application/x-msi",
	"application/x-sdlc", // Oracle
	"application/x-zip-compressed",
	"application/zip",
	"binary/octet-stream",
	"Composite Document File V2 Document, corrupt: Can't read SAT; charset=binary", // Google Code
	"text/x-python", // PIP
	"Zip Files",
}

// FIXME: Latest version of gojsonschema finds errors that it didn't find before.
//
// func TestValidRegistry(t *testing.T) {
// 	schemaLoader := gojsonschema.NewReferenceLoader("file://just-install-schema.json")
// 	documentLoader := gojsonschema.NewReferenceLoader("file://just-install.json")
//
// 	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
//
// 	assert.Nil(t, err)
// 	assert.Empty(t, result.Errors())
// 	assert.True(t, result.Valid())
// }

func TestRegistryReachableLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping reachability test in short mode")
	}

	hasErrors := false
	registry := SmartLoadRegistry(false)

	checkLink := func(rawurl string) error {

		response, err := customGet(rawurl)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			t.Log(rawurl)

			return errors.New("Status code wasn't 200 OK")
		}

		contentType := response.Header.Get("Content-Type")

		// Exception: VirtualBox Extension Pack have the wrong MIME type
		success := strings.HasSuffix(rawurl, ".vbox-extpack") && contentType == "text/plain"

		// Exception: Some LibreOffice mirror returns the wrong MIME type
		success = success || strings.Contains(rawurl, "libreoffice") && contentType == "application/x-troff-man"

		success = success || dry.StringInSlice(contentType, expectedContentTypes)
		if !success {
			t.Log(rawurl)

			return errors.New("The content type was " + contentType)
		}

		return nil
	}

	checkArch := func(name string, entry *registryEntry, architecture string, rawUrl string) {
		if rawUrl == "" {
			return
		}

		url := entry.expandString(rawUrl)
		if err := checkLink(url); err != nil {
			t.Logf("%v (%v): %v, %v", name, architecture, url, err)
			hasErrors = true
		}
	}

	for _, name := range registry.SortedPackageNames() {
		entry := registry.Packages[name]

		checkArch(name, &entry, "x86", entry.Installer.X86)
		checkArch(name, &entry, "x86_64", entry.Installer.X86_64)
	}

	assert.False(t, hasErrors)
}
