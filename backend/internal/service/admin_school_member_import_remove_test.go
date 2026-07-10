package service

import (
	"database/sql/driver"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// sameValueArg is a sqlmock argument matcher that records the first value it sees
// and requires every subsequent match to be equal to it. This is how we assert
// that RemoveMember uses the *same* leftAt timestamp for both the school_users
// soft-delete and the enrollments cascade update.
type sameValueArg struct {
	value *driver.Value
}

func (a *sameValueArg) Match(v driver.Value) bool {
	if *a.value == nil {
		*a.value = v
		return true
	}
	return *a.value == v
}

func newRemoveMemberTestService(t *testing.T) (*adminSchoolMemberImportService, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm on top of sqlmock: %v", err)
	}

	svc := NewAdminSchoolMemberImportService(gormDB, nil).(*adminSchoolMemberImportService)
	return svc, mock, func() { sqlDB.Close() }
}

// TestRemoveMemberCascadesEnrollmentLeftAt verifies the happy path: soft-deleting a
// school_user and closing all of their open enrollments happen inside a single
// transaction and share the exact same leftAt timestamp.
func TestRemoveMemberCascadesEnrollmentLeftAt(t *testing.T) {
	svc, mock, closeDB := newRemoveMemberTestService(t)
	defer closeDB()

	shared := &sameValueArg{value: new(driver.Value)}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "edv"\."school_users" SET .*"deleted_at"=\$1.*WHERE scu_id = \$2 AND scu_sch_id = \$3 AND deleted_at IS NULL`).
		WithArgs(shared, "school-user-1", "school-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "edv"\."enrollments" SET "left_at"=\$1 WHERE enr_scu_id = \$2 AND left_at IS NULL`).
		WithArgs(shared, "school-user-1").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	if err := svc.RemoveMember("school-1", "school-user-1"); err != nil {
		t.Fatalf("RemoveMember returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}

	if got, ok := (*shared.value).(time.Time); !ok {
		t.Fatalf("captured leftAt value is not a time.Time: %#v", *shared.value)
	} else if got.IsZero() {
		t.Fatalf("captured leftAt timestamp is zero")
	}
}

// TestRemoveMemberRollsBackWhenSchoolUserNotFound verifies that when the school_user
// row can't be matched (wrong school, already removed, etc.), the transaction is
// rolled back and the enrollments update never runs.
func TestRemoveMemberRollsBackWhenSchoolUserNotFound(t *testing.T) {
	svc, mock, closeDB := newRemoveMemberTestService(t)
	defer closeDB()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "edv"\."school_users" SET .*"deleted_at"=\$1.*WHERE scu_id = \$2 AND scu_sch_id = \$3 AND deleted_at IS NULL`).
		WithArgs(sqlmock.AnyArg(), "school-user-1", "school-1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	if err := svc.RemoveMember("school-1", "school-user-1"); err == nil {
		t.Fatalf("expected an error when school_user does not match, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

// TestRemoveMemberRollsBackWhenEnrollmentUpdateFails verifies that a failure in the
// second statement (enrollments cascade) rolls back the first statement too, so a
// school_user is never left soft-deleted with its enrollments still open.
func TestRemoveMemberRollsBackWhenEnrollmentUpdateFails(t *testing.T) {
	svc, mock, closeDB := newRemoveMemberTestService(t)
	defer closeDB()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "edv"\."school_users" SET .*"deleted_at"=\$1.*WHERE scu_id = \$2 AND scu_sch_id = \$3 AND deleted_at IS NULL`).
		WithArgs(sqlmock.AnyArg(), "school-user-1", "school-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "edv"\."enrollments" SET "left_at"=\$1 WHERE enr_scu_id = \$2 AND left_at IS NULL`).
		WithArgs(sqlmock.AnyArg(), "school-user-1").
		WillReturnError(sqlmock.ErrCancelled)
	mock.ExpectRollback()

	if err := svc.RemoveMember("school-1", "school-user-1"); err == nil {
		t.Fatalf("expected an error when the enrollment cascade update fails, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
