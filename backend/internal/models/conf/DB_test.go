package conf

import "testing"

func TestDBParse(t *testing.T) {
	t.Run("parses all values from the environment", func(t *testing.T) {
		t.Setenv("DB_HOST", "db.example.com")
		t.Setenv("DB_USER", "avuser")
		t.Setenv("DB_PASSWORD", "s3cret")
		t.Setenv("DB_NAME", "av_directory")
		t.Setenv("DB_PORT", "6543")
		t.Setenv("DB_SSLMODE", "require")

		var db DB
		if err := db.Parse(); err != nil {
			t.Fatalf("Parse() returned an error: %v", err)
		}

		want := DB{
			Host:      "db.example.com",
			User:      "avuser",
			Password:  "s3cret",
			TableName: "av_directory",
			Port:      "6543",
			SSLMode:   "require",
		}
		if db != want {
			t.Errorf("Parse() = %+v, want %+v", db, want)
		}
	})

	t.Run("falls back to defaults when unset", func(t *testing.T) {
		t.Setenv("DB_HOST", "")
		t.Setenv("DB_USER", "")
		t.Setenv("DB_PASSWORD", "")
		t.Setenv("DB_NAME", "")
		t.Setenv("DB_PORT", "")
		t.Setenv("DB_SSLMODE", "")

		var db DB
		if err := db.Parse(); err != nil {
			t.Fatalf("Parse() returned an error: %v", err)
		}

		want := DB{
			Host:      "localhost",
			User:      "postgres",
			Password:  "",
			TableName: "omniav",
			Port:      "5432",
			SSLMode:   "disable",
		}
		if db != want {
			t.Errorf("Parse() = %+v, want %+v", db, want)
		}
	})
}

func TestDBDSN(t *testing.T) {
	db := DB{
		Host:      "db.example.com",
		User:      "avuser",
		Password:  "s3cret",
		TableName: "av_directory",
		Port:      "6543",
		SSLMode:   "require",
	}

	want := "host=db.example.com user=avuser password=s3cret dbname=av_directory port=6543 sslmode=require"
	if got := db.DSN(); got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}
