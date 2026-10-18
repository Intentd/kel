package core

import "testing"

func TestCronFieldExtractors(t *testing.T) {
	tests := []struct {
		name                                              string
		expr                                              string
		wantMinute, wantHour, wantDom, wantMonth, wantDow string
		wantErr                                           bool
	}{
		{"standard 5 fields", "*/5 0 * * 1", "*/5", "0", "*", "*", "1", false},
		{"@hourly macro", "@hourly", "0", "*", "*", "*", "*", false},
		{"@daily macro", "@daily", "0", "0", "*", "*", "*", false},
		{"@weekly macro", "@weekly", "0", "0", "*", "*", "0", false},
		{"@monthly macro", "@monthly", "0", "0", "1", "*", "*", false},
		{"@yearly macro", "@yearly", "0", "0", "1", "1", "*", false},
		{"empty string", "", "*", "*", "*", "*", "*", false},
		{"invalid field count", "0 0 * *", "*", "*", "*", "*", "*", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, err := CronMinute(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("cronMinute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if min != tt.wantMinute {
				t.Errorf("cronMinute() = %q, want %q", min, tt.wantMinute)
			}
			hour, _ := CronHour(tt.expr)
			if hour != tt.wantHour {
				t.Errorf("cronHour() = %q, want %q", hour, tt.wantHour)
			}
			dom, _ := CronDom(tt.expr)
			if dom != tt.wantDom {
				t.Errorf("cronDom() = %q, want %q", dom, tt.wantDom)
			}
			month, _ := CronMonth(tt.expr)
			if month != tt.wantMonth {
				t.Errorf("cronMonth() = %q, want %q", month, tt.wantMonth)
			}
			dow, _ := CronDow(tt.expr)
			if dow != tt.wantDow {
				t.Errorf("cronDow() = %q, want %q", dow, tt.wantDow)
			}
		})
	}
}

func TestCronField(t *testing.T) {
	tests := []struct {
		expr string
		pos  int
		want string
		err  bool
	}{
		{"*/5 0 * * 1", 0, "*/5", false},
		{"*/5 0 * * 1", 1, "0", false},
		{"*/5 0 * * 1", 2, "*", false},
		{"*/5 0 * * 1", 3, "*", false},
		{"*/5 0 * * 1", 4, "1", false},
		{"", 0, "*", false},
		{"@hourly", 0, "0", false},
		{"  invalid count  ", 0, "*", true},
		{"*/5 0 * * 1", 5, "*", true},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := CronField(tt.expr, tt.pos)
			if (err != nil) != tt.err {
				t.Errorf("cronField() error = %v, wantErr %v", err, tt.err)
				return
			}
			if got != tt.want {
				t.Errorf("cronField() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCronExpr(t *testing.T) {
	tests := []struct {
		min, hour, dom, month, dow string
		want                       string
	}{
		{"*/5", "0", "*", "*", "1", "*/5 0 * * 1"},
		{"", "0", "*", "*", "1", "* 0 * * 1"},
		{"*/5", "", "*", "*", "1", "*/5 * * * 1"},
	}
	for _, tt := range tests {
		got := CronExpr(tt.min, tt.hour, tt.dom, tt.month, tt.dow)
		if got != tt.want {
			t.Errorf("cronExpr(%q,%q,%q,%q,%q) = %q, want %q", tt.min, tt.hour, tt.dom, tt.month, tt.dow, got, tt.want)
		}
	}
}

func TestCronValid(t *testing.T) {
	tests := []struct {
		expr string
		want bool
	}{
		{"*/5 0 * * 1", true},
		{"@hourly", true},
		{"", false},
		{"0 0 * *", false},
		{"invalid", false},
	}
	for _, tt := range tests {
		if got := CronValid(tt.expr); got != tt.want {
			t.Errorf("cronValid(%q) = %v, want %v", tt.expr, got, tt.want)
		}
	}
}

func TestCronFromMapStrict(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want string
		err  bool
	}{
		{"valid map", map[string]interface{}{"minute": "*/5", "hour": "0", "dayOfMonth": "*", "month": "*", "dayOfWeek": "1"}, "*/5 0 * * 1", false},
		{"missing fields", map[string]interface{}{"minute": "*/5"}, "*/5 * * * *", false},
		{"not a map", "string", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CronFromMapStrict(tt.v)
			if (err != nil) != tt.err {
				t.Errorf("cronFromMapStrict() error = %v, wantErr %v", err, tt.err)
				return
			}
			if got != tt.want {
				t.Errorf("cronFromMapStrict() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCronFromMapAny(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want string
	}{
		{"map", map[string]interface{}{"minute": "*/5", "hour": "0", "dayOfMonth": "*", "month": "*", "dayOfWeek": "1"}, "*/5 0 * * 1"},
		{"cron string", "*/5 0 * * 1", "*/5 0 * * 1"},
		{"@hourly", "@hourly", "0 * * * *"},
		{"empty", nil, "* * * * *"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CronFromMapAny(tt.v); got != tt.want {
				t.Errorf("cronFromMapAny(%v) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestCronNormalize(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		{"*/5 0 * * 1", "*/5 0 * * 1"},
		{"@hourly", "0 * * * *"},
		{"", "* * * * *"},
		{"0 0 * *", "* * * * *"},
		{"  0 0 * * *  ", "0 0 * * *"},
	}
	for _, tt := range tests {
		if got := CronNormalize(tt.expr); got != tt.want {
			t.Errorf("cronNormalize(%q) = %q, want %q", tt.expr, got, tt.want)
		}
	}
}

func TestCronDescribe(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		{"* * * * *", "Every minute"},
		{"*/5 * * * *", "Every 5 minutes"},
		{"0 * * * *", "Every hour"},
		{"15 * * * *", "At minute 15 of every hour"},
		{"0 */2 * * *", "Every 2 hours"},
		{"30 */6 * * *", "At minute 30, every 6 hours"},
		{"30 2 * * *", "At 02:30 every day"},
		{"@daily", "At 00:00 every day"},
		{"0 2 * * 1", "At 02:00 on Monday"},
		{"0 2 * * 1-5", "At 02:00 on Monday through Friday"},
		{"0 9 * * 1,3,5", "At 09:00 on Monday, Wednesday and Friday"},
		{"0 9 * * 0", "At 09:00 on Sunday"},
		{"0 9 * * SAT", "At 09:00 on Saturday"},
		{"0 2 15 * *", "At 02:00 on day 15 of every month"},
		{"0 0 1 6 *", "At 00:00 on day 1 in June"},
		{"0 0 * 12 *", "At 00:00 every day in December"},
		{"*/15 9-17 * * 1-5", "Cron: */15 9-17 * * 1-5"},
		{"0 2 15 * 1", "Cron: 0 2 15 * 1"},
		{"invalid", "Invalid cron expression"},
	}
	for _, tt := range tests {
		if got := CronDescribe(tt.expr); got != tt.want {
			t.Errorf("cronDescribe(%q) = %q, want %q", tt.expr, got, tt.want)
		}
	}
}
