package email

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"strconv"
	"strings"
)

type SMTPSender struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

func NewSMTPSender(host, portStr, user, password, from string) (*SMTPSender, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 587
	}
	return &SMTPSender{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
	}, nil
}

func (s *SMTPSender) Send(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		s.from, to, subject,
	)
	message := []byte(headers + htmlBody)

	auth := smtp.PlainAuth("", s.user, s.password, s.host)

	if s.port == 587 {
		conn, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("smtp dial: %w", err)
		}
		defer func(conn *smtp.Client) {
			err := conn.Close()
			if err != nil {

			}
		}(conn)

		tlsCfg := &tls.Config{ServerName: s.host}
		if err := conn.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
		if err := conn.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		if err := conn.Mail(s.from); err != nil {
			return fmt.Errorf("smtp mail: %w", err)
		}
		if err := conn.Rcpt(to); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
		wc, err := conn.Data()
		if err != nil {
			return fmt.Errorf("smtp data: %w", err)
		}
		defer func(wc io.WriteCloser) {
			err := wc.Close()
			if err != nil {

			}
		}(wc)
		_, err = wc.Write(message)
		return err
	}

	tlsCfg := &tls.Config{InsecureSkipVerify: false, ServerName: s.host}
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}
	c, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer func(c *smtp.Client) {
		err := c.Close()
		if err != nil {

		}
	}(c)

	if err = c.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err = c.Mail(s.from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err = c.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data writer: %w", err)
	}
	defer func(wc io.WriteCloser) {
		err := wc.Close()
		if err != nil {

		}
	}(wc)
	_, err = wc.Write(message)
	return err
}

func WelcomeEmail(name string) string {
	return fmt.Sprintf(`
<html><body>
<h2>Добро пожаловать в AITUdel, %s! </h2>
<p>Рады видеть вас на нашей платформе. Заказывайте еду быстро и удобно.</p>
<p>С уважением,<br>Команда AITUdel</p>
</body></html>`, name)
}

func StudentWelcomeEmail(name, promoCode string) string {
	return fmt.Sprintf(`
<html><body>
<h2>Добро пожаловать в AITUdel, %s! </h2>
<p>Вы зарегистрированы как студент AITU.</p>
<p>Ваш промокод: <strong>%s</strong> — скидка 15%% на все заказы!</p>
<p>С уважением,<br>Команда AITUdel</p>
</body></html>`, name, promoCode)
}

const orderNamePlaceholder = "{{NAME}}"

func OrderDeliveredEmail(name, orderID string, total float64) string {
	displayName := name
	if displayName == "" {
		displayName = orderNamePlaceholder
	}
	orderDisplay := orderID
	if orderDisplay == "" {
		orderDisplay = "—"
	}
	return fmt.Sprintf(`
<html><body>
<h2>Ваш заказ доставлен, %s! </h2>
<p>Заказ <strong>#%s</strong> успешно доставлен.</p>
<p>Сумма заказа: <strong>%.2f ₸</strong></p>
<p>Приятного аппетита!<br>Команда AITUdel</p>
</body></html>`, displayName, orderDisplay, total)
}

func InjectNameIntoOrderEmail(name, body string) string {
	return strings.ReplaceAll(body, orderNamePlaceholder, name)
}
