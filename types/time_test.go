package types

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestTimeBod(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "BOD", args: args{t: time.Time{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeBod(tt.args.t); got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
				t.Errorf("TimeBod() = %v, want all Hour Minute Second Nanosecond as 0", got)
			}
		})
	}
}

func TestTimeBom(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{name: "BOM", args: args{t: time.Time{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeBom(tt.args.t); got.Day() != 1 || got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
				t.Errorf("TimeBom() = %v, want Day as 1 and  all Hour Minute Second Nanosecond as 0", got)
			}
		})
	}
}

func TestTimeMonSunWeekDay(t *testing.T) {
	mon, _ := time.Parse("01/02/2006", "03/01/2021")
	tue, _ := time.Parse("01/02/2006", "03/02/2021")
	wed, _ := time.Parse("01/02/2006", "03/03/2021")
	thu, _ := time.Parse("01/02/2006", "03/04/2021")
	fri, _ := time.Parse("01/02/2006", "03/05/2021")
	sat, _ := time.Parse("01/02/2006", "03/06/2021")
	sun, _ := time.Parse("01/02/2006", "03/07/2021")
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "mon", args: args{t: mon}, want: 0},
		{name: "tue", args: args{t: tue}, want: 1},
		{name: "wed", args: args{t: wed}, want: 2},
		{name: "thu", args: args{t: thu}, want: 3},
		{name: "fri", args: args{t: fri}, want: 4},
		{name: "sat", args: args{t: sat}, want: 5},
		{name: "sun", args: args{t: sun}, want: 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeMonSunWeekDay(tt.args.t); got != tt.want {
				t.Errorf("TimeMonSunWeekDay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateEqual(t *testing.T) {
	type args struct {
		date1 time.Time
		date2 time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "positive", args: args{date1: time.Time{}, date2: time.Time{}}, want: true},
		{name: "negative", args: args{date1: time.Time{}, date2: time.Time{}.AddDate(0, 0, -1)}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DateEqual(tt.args.date1, tt.args.date2); got != tt.want {
				t.Errorf("DateEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseUnix(t *testing.T) {
	type args struct {
		msString string
	}
	now := time.Time{}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{name: "", args: args{msString: strconv.FormatInt(now.UnixNano()/1000000, 64)}, want: now},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseUnix(tt.args.msString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseUnix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseUnix() = %v, want %v", got, tt.want)
			}
		})
	}
}
