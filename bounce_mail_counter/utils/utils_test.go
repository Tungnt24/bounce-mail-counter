package utils

import (
	"fmt"
	"testing"
)

func TestDetectSpam(t *testing.T) {
	type args struct {
		message string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Spam1",
			args: args{`postfix/smtp[6680]: 1D326DADD9: to=<hr@kamereo.com>, relay=mail.kamereo.com[149.255.60.185]:25, delay=3.6, delays=0.01/0.02/3.3/0.3, dsn=5.1.1, status=bounced (host mail.kamereo.com[149.255.60.185] said: 550 5.1.1 <hr@kamereo.com>: Recipient address rejected: User unknown in virtual mailbox table (in reply to RCPT TO command))`},
			want: true,
		},
		{
			name: "Spam2",
			args: args{"postfix/smtp[19058]: 4E9B6DACA3: to=<naoya-yamashita@ad-vance.co.jp>, relay=vlmx20.secure.ne.jp[122.200.253.221]:25, delay=1.7, delays=0.01/0/1.5/0.26, dsn=5.0.0, status=bounced (host vlmx20.secure.ne.jp[122.200.253.221] said: 550 #5.1.0 Address rejected. (in reply to RCPT TO command))"},
			want: true,
		},
		{
			name: "Spam3",
			args: args{"postfix/smtp[1214283]: 29A5960616: to=<wined54@hotmail.com>, relay=hotmail-com.olc.protection.outlook.com[104.47.6.33]:25, delay=12, delays=0.05/0/1.5/10, dsn=5.7.1, status=bounced (host hotmail-com.olc.protection.outlook.com[104.47.6.33] said: 550 5.7.1 Service unavailable, MailFrom domain is listed in Spamhaus. To request removal from this list see https://www.spamhaus.org/query/lookup/ (S8002) [VE1EUR02FT012.eop-EUR02.prod.protection.outlook.com] (in reply to end of DATA command))"},
			want: true,
		},
		{
			name: "Not Spam1",
			args: args{`postfix/smtp[26753]: 517A7DAE7D: to=<amove@amail.plala.or.jp>, relay=mx.plala.or.jp[60.36.166.235]:25, delay=0.73, delays=0.02/0/0.58/0.12, dsn=5.0.0, status=bounced (host mx.plala.or.jp[60.36.166.235] said: 550  (in reply to RCPT TO command))`},
			want: false,
		},
		{
			name: "Not Spam2",
			args: args{"postfix/smtp[6068]: 8B82CDAE6D: to=<ctphongquang@gmail.com>, relay=gmail-smtp-in.l.google.com[64.233.189.27]:25, delay=1.5, delays=0.01/0/1.3/0.21, dsn=5.1.1, status=bounced (host gmail-smtp-in.l.google.com[64.233.189.27] said: 550-5.1.1 The email account that you tried to reach does not exist. Please try 550-5.1.1 double-checking the recipient's email address for typos or 550-5.1.1 unnecessary spaces. Learn more at 550 5.1.1  https://support.google.com/mail/?p=NoSuchUser k128si6689234pfd.240 - gsmtp (in reply to RCPT TO command))"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Print(tt.args.message)
			if got := DetectSpam(tt.args.message); got != tt.want {
				t.Errorf("DetectSpam() = %v, want %v", got, tt.want)
			}
		})
	}
}
