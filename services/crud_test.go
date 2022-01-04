package services

import (
	"context"
	"reflect"
	"testing"
	"time"

	"gitlab.com/sincap/sincap-common/db"
	"gitlab.com/sincap/sincap-common/repositories"
	"gorm.io/gorm"
)

// User model for the app
type user struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Username  string         `sql:"index" gorm:"index:,unique;not null;size:50" validate:"required,email,min=6,max=50" qapi:"q:*%;"`
}
type repository interface {
	repositories.Repository[user]
}

func Test_CrudService_Read(t *testing.T) {
	mock := db.ConfigureMockDB("Test_CrudService_Read")
	rep := repositories.NewGormRepository[user](mock)
	mock.AutoMigrate(user{})

	rep.Create(&user{Username: "test"})

	type fields struct {
		repository repositories.Repository[user]
	}
	type args struct {
		ctx context.Context
		uid uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{name: "success", fields: fields{repository: &rep}, args: args{ctx: context.Background(), uid: 1}, want: "test"},
		{name: "fail", fields: fields{repository: &rep}, args: args{ctx: context.Background(), uid: 2}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &CrudService[user]{
				Repository: tt.fields.repository,
			}
			got, err := ser.Read(tt.args.ctx, tt.args.uid)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && !reflect.DeepEqual(got.Username, tt.want) {
				t.Errorf("service.Read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_CrudService_Update(t *testing.T) {
	mock := db.ConfigureMockDB("Test_CrudService_Update")
	rep := repositories.NewGormRepository[user](mock)
	mock.AutoMigrate(user{})

	rep.Create(&user{Username: "test"})

	t.Run("success", func(t *testing.T) {
		ser := &CrudService[user]{
			Repository: &rep,
		}
		err := ser.Update(context.Background(), 1, map[string]interface{}{"Username": "test2"})
		if err != nil {
			t.Errorf("service.Update() error = %v", err)
			return
		}
		u := user{}
		if err := ser.Repository.Read(&u, uint(1)); err != nil {
			t.Errorf("service.Update() error = %v", err)
			return
		}
		if u.Username != "test2" {
			t.Errorf("service.Update() error = %v", err)
			return
		}

	})
}
func Test_CrudService_Delete(t *testing.T) {
	mock := db.ConfigureMockDB("Test_CrudService_Delete")
	rep := repositories.NewGormRepository[user](mock)
	mock.AutoMigrate(user{})

	rep.Create(&user{Username: "test"})

	t.Run("success", func(t *testing.T) {
		ser := &CrudService[user]{
			Repository: &rep,
		}
		u, err := ser.Delete(context.Background(), uint(1))
		if err != nil {
			t.Errorf("service.Delete() error = %v", err)
			return
		}
		if err := ser.Repository.Read(u, uint(1)); err == nil {
			t.Errorf("service.Delete() error = %v", err)
			return
		}
	})
}
