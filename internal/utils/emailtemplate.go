package utils

import (
	"bytes"
	"html/template"
	"strconv"
	"time"
)

type EmailTemplateData struct {
	OTPCode       string
	UserName      string
	CompanyName   string
	ExpiryMinutes int32
	Year          string
}

func GetEmailTemplate(otp string, userName string, companyName string, expiryMinutes int32) (string, error) {
	tmpl, err := template.New("otp-email").Parse(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>OTP Email</title>
</head>
<body style="margin:0; padding:0; background-color:#f4f6f8; font-family:Arial, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f6f8; padding:20px 0;">
    <tr>
      <td align="center">
        <table width="400" cellpadding="0" cellspacing="0" style="background:#ffffff; border-radius:8px; padding:30px; box-shadow:0 2px 8px rgba(0,0,0,0.05);">
          <tr>
            <td align="center" style="padding-bottom:20px;">
              <h2 style="margin:0; color:#333;">{{.CompanyName}}</h2>
            </td>
          </tr>
          <tr>
            <td style="color:#555; font-size:14px; line-height:1.6;">
              Hi {{.UserName}},<br><br>
              Your One-Time Password (OTP) is:
            </td>
          </tr>
          <tr>
            <td align="center" style="padding:20px 0;">
              <div style="display:inline-block; padding:15px 25px; font-size:24px; font-weight:bold; letter-spacing:4px; color:#000; background:#f1f3f5; border-radius:6px;">
                {{.OTPCode}}
              </div>
            </td>
          </tr>
          <tr>
            <td style="color:#555; font-size:14px;">
              This code will expire in <strong>{{.ExpiryMinutes}} minutes</strong>.
            </td>
          </tr>
          <tr>
            <td style="color:#aaa; font-size:12px; padding-top:25px; text-align:center;">
              If you didn’t request this, you can safely ignore this email.
              <br><br>
              © {{.Year}} {{.CompanyName}}. All rights reserved.
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`)
	if err != nil {
		return "", err
	}

	data := EmailTemplateData{
		OTPCode:       otp,
		UserName:      userName,
		CompanyName:   companyName,
		ExpiryMinutes: expiryMinutes,
		Year:          strconv.Itoa(time.Now().Year()), // or time.Now().Year()
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
