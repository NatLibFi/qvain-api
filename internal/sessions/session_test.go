package sessions

import (
	"bytes"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/CSCfi/qvain-api/pkg/models"
	"github.com/gomodule/redigo/redis"
	"github.com/wvh/uuid"
)

func TestSessionManager(t *testing.T) {
	mgr := NewManager()
	user1 := &models.User{
		Uid:          uuid.MustFromString("053bffbcc41edad4853bea91fc42ea18"),
		Identity:     "identity1@oidc",
		Name:         "User One",
		Email:        "one@example.com",
		Organisation: "Test Organisation",
		Projects:     []string{"project1", "project2", "project3"},
	}

	user2 := &models.User{
		Uid:          uuid.MustFromString("053bffbcc41edad4853bea91fc42ea19"),
		Identity:     "identity2@oidc",
		Name:         "User Two",
		Email:        "two@example.com",
		Organisation: "Test Organisation",
		Projects:     []string{"project1"},
	}

	t.Run("addOne", func(t *testing.T) {
		mgr.new("sid-for-one", &user1.Uid, user1)
		if !mgr.Exists("sid-for-one") {
			t.Error("session `one` should exist")
		}
	})
	t.Run("addTwoAndCount", func(t *testing.T) {
		mgr.new("sid-for-two", &user2.Uid, user2)
		count := mgr.Count()
		if count != 2 {
			t.Error("there should be two sessions, got:", count)
		}
	})
	t.Run("addThreeWithoutUid", func(t *testing.T) {
		if mgr.new("sid-for-three-without-uid", nil, nil) != nil {
			t.Error("can't add session with nil uid")
		}
	})
	t.Run("GetOne", func(t *testing.T) {
		session, err := mgr.Get("sid-for-one")
		if err != nil {
			t.Error("should return session for sid `sid-for-one`, got:", err)
		}
		if session.User.Identity != "identity1@oidc" {
			t.Errorf("should return identity %q, got %q", user1.Identity, session.User.Identity)
		}
		uid, err := session.Uid()
		if err != nil {
			t.Errorf("should return UUID uid, got uid %v, error: %s", uid, err)
		}
		//t.Logf("uid: %v", uid)
	})
	t.Run("GetOneWithUid", func(t *testing.T) {
		session, err := mgr.Get("sid-for-three-without-uid")
		if err != nil {
			t.Error("should return session for sid `sid-for-three-without-uid`, got:", err)
		}
		uid, err := session.Uid()
		if err != ErrUnknownUser {
			t.Errorf("should return nil uid with error, got uid %v, error: %s", uid, err)
		}
	})
	t.Run("GetNonExisting", func(t *testing.T) {
		session, err := mgr.Get("nonexisting")
		if err == nil {
			t.Error("should get ErrSessionNotFound error but got: nil")
		}
		if err != ErrSessionNotFound {
			t.Error("should get ErrSessionNotFound error but got:", err)
		}
		if session != nil {
			t.Errorf("should return nil session if it doesn't exist, got: %v", session)
		}
	})
	t.Run("DestroyNonExisting", func(t *testing.T) {
		if mgr.Destroy("nonexisting") {
			t.Error("should return false if session didn't exist (anymore), got: true")
		}
	})
	t.Run("DestroyTwo", func(t *testing.T) {
		if !mgr.Destroy("sid-for-two") {
			t.Error("should have destroyed session `sid-for-two` but got: false")
		}
		count := mgr.Count()
		if count != 2 {
			t.Error("there should be only two sessions left, got:", count)
		}
	})
	t.Run("SessionWithProjects", func(t *testing.T) {
		if !mgr.Exists("sid-for-one") {
			t.Error("session `sid-for-one` should exist")
		}
		session, err := mgr.Get("sid-for-one")
		if err != nil {
			t.Error("should return session `sid-for-one`, got:", err)
		}
		projects := session.User.Projects
		if len(projects) < 3 {
			t.Error("should return 3 projects, got:", len(projects))
		}
		if projects[2] != "project3" {
			t.Error("should return `project3` for the third project, got:", projects[2])
		}
		if !session.User.HasProject("project2") {
			t.Error("should return true for HasProject `project2`, got: false")
		}
		if session.User.HasProject("projectNonexisting") {
			t.Error("should return false for non-existing project `projectNonexisting`, got: true")
		}
	})
	t.Run("List", func(t *testing.T) {
		var buf bytes.Buffer
		mgr.List(&buf)
		t.Logf("%s", buf.String())
	})
	t.Run("Export", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}
		mgr.Save()
	})
}

func TestSerialisation(t *testing.T) {
	var (
		testExpiration       time.Time = time.Now().Add(1 * time.Minute).Round(time.Second)
		testExpirationString string    = strconv.FormatInt(testExpiration.Unix(), 10)
	)
	var tests = []struct {
		name    string
		session *Session
		json    string
	}{
		{
			name: "basic",
			session: &Session{
				uid: func(uid uuid.UUID) *uuid.UUID {
					return &uid
				}(uuid.MustFromString("053bffbcc41edad4853bea91fc42ea18")),
				User: &models.User{
					Uid:          uuid.MustFromString("053bffbcc41edad4853bea91fc42ea18"),
					Identity:     "identity1@oidc",
					Name:         "User One",
					Email:        "one@example.com",
					Organisation: "Test Organisation",
					Projects:     []string{"project1", "project2", "project3"},
				},
				Expiration: testExpiration,
			},
			json: `{"uid":"053bffbcc41edad4853bea91fc42ea18","expiration":` + testExpirationString + `,"user":{"uid":"053bffbcc41edad4853bea91fc42ea18","identity":"identity1@oidc","name":"User One","email":"one@example.com","organisation":"Test Organisation","projects":["project1","project2","project3"]}}`,
		},
		{
			name: "no projects",
			session: &Session{
				uid: func(uid uuid.UUID) *uuid.UUID {
					return &uid
				}(uuid.MustFromString("053bffbcc41edad4853bea91fc42ea18")),
				User: &models.User{
					Uid:          uuid.MustFromString("053bffbcc41edad4853bea91fc42ea18"),
					Identity:     "identity1@oidc",
					Name:         "User One",
					Email:        "one@example.com",
					Organisation: "Test Organisation",
					// this works in practice but the test fails because nil slice != empty slice, so skip testing empty slice
					//Projects:     []string{},
				},
				Expiration: testExpiration,
			},
			json: `{"uid":"053bffbcc41edad4853bea91fc42ea18","expiration":` + testExpirationString + `,"user":{"uid":"053bffbcc41edad4853bea91fc42ea18","identity":"identity1@oidc","name":"User One","email":"one@example.com","organisation":"Test Organisation"}}`,
		},
		{
			name: "nil uid and user projects",
			session: &Session{
				uid: nil,
				User: &models.User{
					Uid:          uuid.MustFromString("053bffbcc41edad4853bea91fc42ea18"),
					Identity:     "identity1@oidc",
					Name:         "User One",
					Email:        "one@example.com",
					Organisation: "Test Organisation",
					//Projects:     nil,
				},
				Expiration: testExpiration,
			},
			json: `{"uid":null,"expiration":` + testExpirationString + `,"user":{"uid":"053bffbcc41edad4853bea91fc42ea18","identity":"identity1@oidc","name":"User One","email":"one@example.com","organisation":"Test Organisation"}}`,
		},
		// the next test only passes if a nil pointer user is set to encode to a JSON null
		/*
			{
				name: "nil user (to null)",
				session: &Session{
					uid: func(uid uuid.UUID) *uuid.UUID {
						return &uid
					}(uuid.MustFromString("053bffbcc41edad4853bea91fc42ea18")),
					User:       nil,
					Expiration: testExpiration,
				},
				json: `{"uid":"053bffbcc41edad4853bea91fc42ea18","expiration":` + testExpirationString + `,"user":null}`,
			},
		*/
		// the next test only passes if a nil pointer user is omitted instead of serialising to a JSON null
		{
			name: "nil user (omitted)",
			session: &Session{
				uid: func(uid uuid.UUID) *uuid.UUID {
					return &uid
				}(uuid.MustFromString("053bffbcc41edad4853bea91fc42ea18")),
				User:       nil,
				Expiration: testExpiration,
			},
			json: `{"uid":"053bffbcc41edad4853bea91fc42ea18","expiration":` + testExpirationString + `}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name+"(ToJson)", func(t *testing.T) {
			//json, err := gojay.Marshal(test.user)
			json, err := test.session.AsJson()
			if err != nil {
				t.Error("error serialising session:", err)
			}
			if string(json) != test.json {
				t.Errorf("serialisation error:\n\texpected:\n\t\t%s\n\tgot:\n\t\t%s\n", test.json, json)
			}
		})
		t.Run(test.name+"(FromJson)", func(t *testing.T) {
			session := new(Session)
			err := FromJson([]byte(test.json), session)
			if err != nil {
				t.Error("deserialisation error:", err)
			}

			if !reflect.DeepEqual(session, test.session) {
				t.Errorf("sessions don't match after deserialisation:\n\texpected:\n\t\t%+v\n\tgot:\n\t\t%+v\n", test.session, session)
				t.Logf("user:\n\t%+v\n\t\n\t%+v\n", test.session.User, session.User)
			}
		})
	}
}

func TestGetJwtSignature(t *testing.T) {
	var tests = []struct {
		jwt string
		sig string
		err error
	}{
		{
			jwt: "",
			sig: "",
			err: ErrMalformedToken,
		},
		{
			jwt: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			sig: "jwt:SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			err: nil,
		},
		{
			jwt: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c.extra",
			sig: "",
			err: ErrMalformedToken,
		},
		{
			jwt: "SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			sig: "",
			err: ErrMalformedToken,
		},
		{
			// HS256
			jwt: "eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJqb2UiLA0KICJleHAiOjEzMDA4MTkzODAsDQogImh0dHA6Ly9leGFtcGxlLmNvbS9pc19yb290Ijp0cnVlfQ.dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			sig: "jwt:dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			err: nil,
		},
		{
			// RS256
			jwt: "eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJqb2UiLA0KICJleHAiOjEzMDA4MTkzODAsDQogImh0dHA6Ly9leGFtcGxlLmNvbS9pc19yb290Ijp0cnVlfQ.cC4hiUPoj9Eetdgtv3hF80EGrhuB__dzERat0XF9g2VtQgr9PJbu3XOiZj5RZmh7AAuHIm4Bh-0Qc_lF5YKt_O8W2Fp5jujGbds9uJdbF9CUAr7t1dnZcAcQjbKBYNX4BAynRFdiuB--f_nZLgrnbyTyWzO75vRK5h6xBArLIARNPvkSjtQBMHlb1L07Qe7K0GarZRmB_eSN9383LcOLn6_dO--xi12jzDwusC-eOkHWEsqtFZESc6BfI7noOPqvhJ1phCnvWh6IeYI2w9QOYEUipUTI8np6LbgGY9Fs98rqVt5AXLIhWkWywlVmtVrBp0igcN_IoypGlUPQGe77Rw",
			//sig: "jwt:cC4hiUPoj9Eetdgtv3hF80EGrhuB__dzERat0XF9g2VtQgr9PJbu3XOiZj5RZmh7AAuHIm4Bh-0Qc_lF5YKt_O8W2Fp5jujGbds9uJdbF9CUAr7t1dnZcAcQjbKBYNX4BAynRFdiuB--f_nZLgrnbyTyWzO75vRK5h6xBArLIARNPvkSjtQBMHlb1L07Qe7K0GarZRmB_eSN9383LcOLn6_dO--xi12jzDwusC-eOkHWEsqtFZESc6BfI7noOPqvhJ1phCnvWh6IeYI2w9QOYEUipUTI8np6LbgGY9Fs98rqVt5AXLIhWkWywlVmtVrBp0igcN_IoypGlUPQGe77Rw",
			sig: "",
			err: ErrMalformedToken,
		},
		{
			// ES256
			jwt: "eyJhbGciOiJFUzI1NiJ9.eyJpc3MiOiJqb2UiLA0KICJleHAiOjEzMDA4MTkzODAsDQogImh0dHA6Ly9leGFtcGxlLmNvbS9pc19yb290Ijp0cnVlfQ.DtEhU3ljbEg8L38VWAfUAqOyKAM6-Xx-F4GawxaepmXFCgfTjDxw5djxLa8ISlSApmWQxfKTUJqPP3-Kg6NU1Q",
			sig: "jwt:DtEhU3ljbEg8L38VWAfUAqOyKAM6-Xx-F4GawxaepmXFCgfTjDxw5djxLa8ISlSApmWQxfKTUJqPP3-Kg6NU1Q",
			err: nil,
		},
	}

	for _, test := range tests {
		sig, err := GetJwtSignature(test.jwt)
		if err == nil && sig != test.sig {
			t.Errorf("signature failed, expected: %q, got: %q", test.sig, sig)
		}
		if err != test.err {
			t.Errorf("signature failed, expected: %q, got: %q", test.err, err)
		}
	}
}

func BenchmarkAddSession(b *testing.B) {
	mgr := NewManager()
	prefix := "abcdefghijklmnopqrstuvwxyz"
	uid, _ := uuid.NewUUID()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		i := strconv.Itoa(n)
		mgr.new(prefix+i, &uid, nil)
	}
}

func BenchmarkAddSessionRedis(b *testing.B) {
	conn, err := redis.Dial("unix", "/home/wouter/.q.redis.sock")
	if err != nil {
		panic(err)
	}
	exp := int64(DefaultExpiration / time.Second)
	prefix := "abcdefghijklmnopqrstuvwxyz"

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		i := strconv.Itoa(n)
		conn.Send("SETEX", prefix+i, exp, i+"@oidc")
		conn.Flush()
	}
}

func BenchmarkGetSession(b *testing.B) {
	mgr := NewManager()
	uid, _ := uuid.NewUUID()
	mgr.new("xxx", &uid, nil)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := mgr.Get("xxx")
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkGetSessionRedis(b *testing.B) {
	mgr := NewManager()
	_ = mgr

	conn, err := redis.Dial("unix", "/home/wouter/.q.redis.sock")
	//conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		panic(err)
	}
	exp := int64(DefaultExpiration / time.Second)
	conn.Send("SETEX", "xxx", exp, "xxx@oidc")
	conn.Flush()
	v, err := conn.Receive()
	if err != nil {
		b.Error("Receive(): error pre-bench:", err)
	}
	b.Logf("Receive(): %v\n", v)
	conn.Do("")

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		//_, err := mgr.GetRedis(conn, "xxx")
		/*
			_, err := conn.Do("GET", "xxx")
			if err != nil {
				b.Error(err)
			}
		*/
		conn.Send("GET", "xxx")
		conn.Flush()
		_, err := conn.Receive()
		if err != nil {
			b.Error(err)
		}
	}
}
