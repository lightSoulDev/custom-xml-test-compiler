package xmlParser_test

import (
	"testing"

	"github.com/lightSoulDev/pixi-xml-test-compiler/internal/xmlParser"
	"github.com/stretchr/testify/assert"
)

func TestXmlParser_ResolveModulePath(t *testing.T) {

	config := xmlParser.NewConfig()
	x := xmlParser.New(config)

	type payload struct {
		id   string
		path string
	}

	testCases := []struct {
		name        string
		payload     payload
		expectError bool
	}{
		{
			name: "common.xml",
			payload: payload{
				id:   "common/TestName",
				path: config.CommonPath + "/common.xml",
			},
			expectError: false,
		},
		{
			name: "common child folder file",
			payload: payload{
				id:   "common/folder/file/TestName",
				path: config.CommonPath + "/folder/file.xml",
			},
			expectError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			modulePath, err := x.ResolveModulePath(testCase.payload.id)

			if testCase.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.payload.path, modulePath)
		})
	}
}
