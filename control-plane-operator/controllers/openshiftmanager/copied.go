package openshiftmanager

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// copied from library-go/manifestclient
var localScheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(localScheme)

func decodeIndividualObj(content []byte) (*unstructured.Unstructured, error) {
	obj, _, err := codecs.UniversalDecoder().Decode(content, nil, &unstructured.Unstructured{})
	if err != nil {
		return nil, fmt.Errorf("unable to decode: %w", err)
	}
	return obj.(*unstructured.Unstructured), nil
}

// TODO: i've changed json.MarshalIndent to json.Marshal
func serializeIndividualObjToJSON(obj *unstructured.Unstructured) (string, error) {
	ret, err := json.Marshal(obj.Object)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

// end of copied from library-go/manifestclient
