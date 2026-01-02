package svd2db

import (
	"os"
	"testing"
)

func TestCreateDB(t *testing.T) {
	fn := "testdata/testdbfile.db"
	os.Remove(fn)
	db, err := db_createdb(fn)
    if db == nil || err != nil {
        t.Errorf(`db_createdb("testdbfile.db") = %v, %v, want !nil, nil`, db, err)
    }
}

func TestCreateDBExistingFile(t *testing.T) {
	fn := "testdata/testdbfile.db"
	db, err := db_createdb(fn)
    if db != nil || err == nil {
        t.Errorf(`db_createdb() = %v, %v, want nil, error`, db, err)
    }
}

func TestInsert(t *testing.T) {
	fn := "testdata/testdbfile.db"
	os.Remove(fn)
	db, err := db_createdb(fn)

	m := map[string]any{"name": "mpu1", "description": "this is mpu1"}
	mpu_id, err := db_insert(db, "mpus", m)

    if mpu_id != 1 || err != nil {
        t.Errorf(`db_insert("name": "mpu1", "description": "this is mpu1") = %v, %v, want 1, nil`, mpu_id, err)
    }

}
