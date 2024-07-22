package api

import (
	"net/http"
	"strings"

	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
)

var payloads = []string{"/graphql", "/api", "/api/graphql", "/graphql/api", "/graphql/graphql"}

var introspect = `{"query": "query IntrospectionQuery{__schema{queryType{name}mutationType{name}subscriptionType{name}types{...FullType}directives{name description locations args{...InputValue}}}}fragment FullType on __Type{kind name description fields(includeDeprecated:true){name description args{...InputValue}type{...TypeRef}isDeprecated deprecationReason}inputFields{...InputValue}interfaces{...TypeRef}enumValues(includeDeprecated:true){name description isDeprecated deprecationReason}possibleTypes{...TypeRef}}fragment InputValue on __InputValue{name description type{...TypeRef}defaultValue}fragment TypeRef on __Type{kind name ofType{kind name ofType{kind name ofType{kind name ofType{kind name ofType{kind name ofType{kind name ofType{kind name}}}}}}}}"}`

func CheckGraphql(urlStr string) {
	target := strings.TrimRight(urlStr, "/")
	for _, payload := range payloads {
		testURL := target + payload
		req, err := http.NewRequest("GET", testURL, nil)
		if err != nil {
			logger.Warnf("Error creating request: %v", err)
			continue
		}
		_, err = checker.CheckServerCustom(req, clients.NoRedirectClient)
		if err != nil {
			continue
		}
		checkIntrospect(testURL)
	}
}

func checkIntrospect(urlStr string) {
	logger.Infof(introspect)
}
