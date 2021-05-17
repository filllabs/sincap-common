package types

import (
	"testing"
	"time"
)

func TestGetMonSunWeekDay(t *testing.T) {
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
				t.Errorf("GetMonSunWeekDay() = %v, want %v", got, tt.want)
			}
		})
	}
}
