package salesforce

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

func Test_validateAuth(t *testing.T) {
	type args struct {
		sf Salesforce
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "validation_success",
			args: args{
				sf: Salesforce{auth: &authentication{
					AccessToken: "1234",
				}},
			},
			wantErr: false,
		},
		{
			name: "validation_fail",
			args: args{
				sf: Salesforce{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateAuth(tt.args.sf); (err != nil) != tt.wantErr {
				t.Errorf("validateAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_usernamePasswordFlow(t *testing.T) {
	auth := authentication{
		AccessToken: "1234",
		InstanceUrl: "example.com",
		Id:          "123abc",
		IssuedAt:    "01/01/1970",
		Signature:   "signed",
		grantType:   grantTypeUsernamePassword,
	}
	server, _ := setupTestServer(auth, http.StatusOK)
	defer server.Close()

	badServer, _ := setupTestServer(auth, http.StatusForbidden)
	defer badServer.Close()

	type args struct {
		domain         string
		username       string
		password       string
		securityToken  string
		consumerKey    string
		consumerSecret string
	}
	tests := []struct {
		name    string
		args    args
		want    *authentication
		wantErr bool
	}{
		{
			name: "authentication_success",
			args: args{
				domain:         server.URL,
				username:       "u",
				password:       "p",
				securityToken:  "t",
				consumerKey:    "key",
				consumerSecret: "secret",
			},
			want:    &auth,
			wantErr: false,
		},
		{
			name: "authentication_fail",
			args: args{
				domain:         badServer.URL,
				username:       "u",
				password:       "p",
				securityToken:  "t",
				consumerKey:    "key",
				consumerSecret: "secret",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := usernamePasswordFlow(tt.args.domain, tt.args.username, tt.args.password, tt.args.securityToken, tt.args.consumerKey, tt.args.consumerSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("loginPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loginPassword() = %v, want %v", *got, *tt.want)
			}
		})
	}
}

func Test_clientCredentialsFlow(t *testing.T) {
	auth := authentication{
		AccessToken: "1234",
		InstanceUrl: "example.com",
		Id:          "123abc",
		IssuedAt:    "01/01/1970",
		Signature:   "signed",
		grantType:   grantTypeClientCredentials,
	}
	server, _ := setupTestServer(auth, http.StatusOK)
	defer server.Close()

	badServer, _ := setupTestServer(auth, http.StatusForbidden)
	defer badServer.Close()

	type args struct {
		domain         string
		consumerKey    string
		consumerSecret string
	}
	tests := []struct {
		name    string
		args    args
		want    *authentication
		wantErr bool
	}{
		{
			name: "authentication_success",
			args: args{
				domain:         server.URL,
				consumerKey:    "key",
				consumerSecret: "secret",
			},
			want:    &auth,
			wantErr: false,
		},
		{
			name: "authentication_fail",
			args: args{
				domain:         badServer.URL,
				consumerKey:    "key",
				consumerSecret: "secret",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := clientCredentialsFlow(tt.args.domain, tt.args.consumerKey, tt.args.consumerSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("clientCredentialsFlow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("clientCredentialsFlow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setAccessToken(t *testing.T) {
	auth := authentication{
		InstanceUrl: "example.com",
		AccessToken: "1234",
	}
	server, _ := setupTestServer(auth, http.StatusOK)
	defer server.Close()

	badServer, _ := setupTestServer(auth, http.StatusForbidden)
	defer badServer.Close()

	type args struct {
		domain      string
		accessToken string
	}
	tests := []struct {
		name    string
		args    args
		want    *authentication
		wantErr bool
	}{
		{
			name: "authentication_success",
			args: args{
				domain:      server.URL,
				accessToken: "1234",
			},
			want:    &auth,
			wantErr: false,
		},
		{
			name: "authentication_fail_http_error",
			args: args{
				domain:      badServer.URL,
				accessToken: "1234",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "authentication_fail_no_token",
			args: args{
				domain:      server.URL,
				accessToken: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setAccessToken(tt.args.domain, tt.args.accessToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("setAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (tt.want == nil && !reflect.DeepEqual(got, tt.want)) ||
				(tt.want != nil && !reflect.DeepEqual(got.AccessToken, tt.want.AccessToken)) {
				t.Errorf("setAccessToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_refreshSession(t *testing.T) {
	refreshedAuth := authentication{
		AccessToken: "1234",
		InstanceUrl: "example.com",
		Id:          "123abc",
		IssuedAt:    "01/01/1970",
		Signature:   "signed",
	}
	serverClientCredentials, sfAuthClientCredentials := setupTestServer(refreshedAuth, http.StatusOK)
	defer serverClientCredentials.Close()
	sfAuthClientCredentials.grantType = grantTypeClientCredentials

	serverUserNamePassword, sfAuthUserNamePassword := setupTestServer(refreshedAuth, http.StatusOK)
	defer serverUserNamePassword.Close()
	sfAuthUserNamePassword.grantType = grantTypeUsernamePassword

	serverNoGrantType, sfAuthNoGrantType := setupTestServer(refreshedAuth, http.StatusOK)
	defer serverNoGrantType.Close()

	serverBadRequest, sfAuthBadRequest := setupTestServer("", http.StatusBadGateway)
	defer serverBadRequest.Close()
	sfAuthBadRequest.grantType = grantTypeClientCredentials

	serverNoRefresh, sfAuthNoRefresh := setupTestServer("", http.StatusOK)
	defer serverNoRefresh.Close()
	sfAuthNoRefresh.grantType = grantTypeClientCredentials

	type args struct {
		auth *authentication
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "refresh_client_credentials",
			args:    args{auth: &sfAuthClientCredentials},
			wantErr: false,
		},
		{
			name:    "refresh_username_password",
			args:    args{auth: &sfAuthUserNamePassword},
			wantErr: false,
		},
		{
			name:    "error_no_grant_type",
			args:    args{auth: &sfAuthNoGrantType},
			wantErr: true,
		},
		{
			name:    "error_bad_request",
			args:    args{auth: &sfAuthBadRequest},
			wantErr: true,
		},
		{
			name:    "no_refresh",
			args:    args{auth: &sfAuthNoRefresh},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := refreshSession(tt.args.auth); (err != nil) != tt.wantErr {
				t.Errorf("refreshSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_jwtFlow(t *testing.T) {
	auth := authentication{
		AccessToken: "1234",
		InstanceUrl: "example.com",
		Id:          "123abc",
		IssuedAt:    "01/01/1970",
		Signature:   "signed",
		grantType:   grantTypeJWT,
	}
	server, _ := setupTestServer(auth, http.StatusOK)
	defer server.Close()

	badServer, _ := setupTestServer(auth, http.StatusForbidden)
	defer badServer.Close()

	type args struct {
		domain         string
		username       string
		consumerKey    string
		consumerRSAPem string
	}
	tests := []struct {
		name    string
		args    args
		want    *authentication
		wantErr bool
	}{
		{
			name: "authentication_success",
			args: args{
				domain:         server.URL,
				username:       "user",
				consumerKey:    "key",
				consumerRSAPem: "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEA4f5wg5l2hKsTeNem/V41fGnJm6gOdrj8ym3rFkEU/wT8RDtn\nSgFEZOQpHEgQ7JL38xUfU0Y3g6aYw9QT0hJ7mCpz9Er5qLaMXJwZxzHzAahlfA0i\ncqabvJOMvQtzD6uQv6wPEyZtDTWiQi9AXwBpHssPnpYGIn20ZZuNlX2BrClciHhC\nPUIIZOQn/MmqTD31jSyjoQoV7MhhMTATKJx2XrHhR+1DcKJzQBSTAGnpYVaqpsAR\nap+nwRipr3nUTuxyGohBTSmjJ2usSeQXHI3bODIRe1AuTyHceAbewn8b462yEWKA\nRdpd9AjQW5SIVPfdsz5B6GlYQ5LdYKtznTuy7wIDAQABAoIBAQCwia1k7+2oZ2d3\nn6agCAbqIE1QXfCmh41ZqJHbOY3oRQG3X1wpcGH4Gk+O+zDVTV2JszdcOt7E5dAy\nMaomETAhRxB7hlIOnEN7WKm+dGNrKRvV0wDU5ReFMRHg31/Lnu8c+5BvGjZX+ky9\nPOIhFFYJqwCRlopGSUIxmVj5rSgtzk3iWOQXr+ah1bjEXvlxDOWkHN6YfpV5ThdE\nKdBIPGEVqa63r9n2h+qazKrtiRqJqGnOrHzOECYbRFYhexsNFz7YT02xdfSHn7gM\nIvabDDP/Qp0PjE1jdouiMaFHYnLBbgvlnZW9yuVf/rpXTUq/njxIXMmvmEyyvSDn\nFcFikB8pAoGBAPF77hK4m3/rdGT7X8a/gwvZ2R121aBcdPwEaUhvj/36dx596zvY\nmEOjrWfZhF083/nYWE2kVquj2wjs+otCLfifEEgXcVPTnEOPO9Zg3uNSL0nNQghj\nFuD3iGLTUBCtM66oTe0jLSslHe8gLGEQqyMzHOzYxNqibxcOZIe8Qt0NAoGBAO+U\nI5+XWjWEgDmvyC3TrOSf/KCGjtu0TSv30ipv27bDLMrpvPmD/5lpptTFwcxvVhCs\n2b+chCjlghFSWFbBULBrfci2FtliClOVMYrlNBdUSJhf3aYSG2Doe6Bgt1n2CpNn\n/iu37Y3NfemZBJA7hNl4dYe+f+uzM87cdQ214+jrAoGAXA0XxX8ll2+ToOLJsaNT\nOvNB9h9Uc5qK5X5w+7G7O998BN2PC/MWp8H+2fVqpXgNENpNXttkRm1hk1dych86\nEunfdPuqsX+as44oCyJGFHVBnWpm33eWQw9YqANRI+pCJzP08I5WK3osnPiwshd+\nhR54yjgfYhBFNI7B95PmEQkCgYBzFSz7h1+s34Ycr8SvxsOBWxymG5zaCsUbPsL0\n4aCgLScCHb9J+E86aVbbVFdglYa5Id7DPTL61ixhl7WZjujspeXZGSbmq0Kcnckb\nmDgqkLECiOJW2NHP/j0McAkDLL4tysF8TLDO8gvuvzNC+WQ6drO2ThrypLVZQ+ry\neBIPmwKBgEZxhqa0gVvHQG/7Od69KWj4eJP28kq13RhKay8JOoN0vPmspXJo1HY3\nCKuHRG+AP579dncdUnOMvfXOtkdM4vk0+hWASBQzM9xzVcztCa+koAugjVaLS9A+\n9uQoqEeVNTckxx0S2bYevRy7hGQmUJTyQm3j1zEUR5jpdbL83Fbq\n-----END RSA PRIVATE KEY-----",
			},
			want:    &auth,
			wantErr: false,
		},
		{
			name: "authentication_fail",
			args: args{
				domain:         badServer.URL,
				username:       "user",
				consumerKey:    "key",
				consumerRSAPem: "-----BEGIN RSA PRIVATE KEY-----\n\n-----END RSA PRIVATE KEY-----",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jwtFlow(tt.args.domain, tt.args.username, tt.args.consumerKey, tt.args.consumerRSAPem, 1*time.Minute)
			if (err != nil) != tt.wantErr {
				t.Errorf("jwtFlow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("jwtFlow() = %v, want %v", got, tt.want)
			}
		})
	}
}
