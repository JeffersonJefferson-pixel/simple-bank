package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	mockdb "simplebank/db/mock"
	db "simplebank/db/sqlc"
	"simplebank/util"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword

	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v)", e.arg, e.password)
}

func EqCreateUserParamsMatcher(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{
		arg,
		password,
	}
}

func TestCreateUserApi(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username":  user.Username,
				"full_name": user.FullName,
				"email":     user.Email,
				"password":  password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParamsMatcher(arg, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			// marhsall body to json
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))

			server.router.ServeHTTP(recorder, request)

			// verify response
			tc.checkResponse(t, recorder)
		})
	}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	return
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, expectedUser db.User) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var actualUser db.User
	err = json.Unmarshal(data, &actualUser)
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, expectedUser.Username, actualUser.Username)
	require.Equal(t, expectedUser.FullName, actualUser.FullName)
	require.Equal(t, expectedUser.Email, actualUser.Email)
	require.NotEmpty(t, expectedUser.HashedPassword)
}
