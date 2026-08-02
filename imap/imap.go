package imap

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"
	"github.com/gophish/gophish/dialer"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"

	"github.com/jordan-wright/email"
)

const (
	maxRawMessageBytes = 1 << 20
	maxHeaderBytes     = 16 << 10
)

var (
	ErrMailboxRecreated       = errors.New("mailbox UID validity changed")
	ErrMessageTooLarge        = errors.New("message exceeds size limit")
	ErrMessageHeadersTooLarge = errors.New("message headers exceed size limit")
)

// Client interface for IMAP interactions
type Client interface {
	Login(username, password string) (cmd *imap.Command, err error)
	Logout(timeout time.Duration) (cmd *imap.Command, err error)
	Select(name string, readOnly bool) (mbox *imap.MailboxStatus, err error)
	Store(seq *imap.SeqSet, item imap.StoreItem, value interface{}, ch chan *imap.Message) (err error)
	Fetch(seqset *imap.SeqSet, items []imap.FetchItem, ch chan *imap.Message) (err error)
}

// Email represents an email.Email with an included IMAP Sequence Number
type Email struct {
	UID         uint32 `json:"uid"`
	UIDValidity uint32 `json:"uid_validity"`
	*email.Email
}

// Mailbox holds onto the credentials and other information
// needed for connecting to an IMAP server.
type Mailbox struct {
	Host             string
	TLS              bool
	IgnoreCertErrors bool
	User             string
	Pwd              string
	Folder           string
	// Read only mode, false (original logic) if not initialized
	ReadOnly bool
}

// Validate validates supplied IMAP model by connecting to the server
func Validate(s *models.IMAP) error {
	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}

	s.Host = s.Host + ":" + strconv.Itoa(int(s.Port)) // Append port
	mailServer := Mailbox{
		Host:             s.Host,
		TLS:              s.TLS,
		IgnoreCertErrors: s.IgnoreCertErrors,
		User:             s.Username,
		Pwd:              s.Password,
		Folder:           s.Folder}

	imapClient, _, err := mailServer.newClient()
	if err != nil {
		log.Error(err.Error())
	} else {
		imapClient.Logout()
	}
	return err
}

func (mbox *Mailbox) MarkAsUnread(emails []Email) error {
	imapClient, status, err := mbox.newClient()
	if err != nil {
		return err
	}
	defer imapClient.Logout()
	seqSet, err := uidSet(emails, status.UidValidity)
	if err != nil {
		return err
	}
	item := imap.FormatFlagsOp(imap.RemoveFlags, true)
	err = imapClient.UidStore(seqSet, item, imap.SeenFlag, nil)
	if err != nil {
		return err
	}

	return nil

}

func (mbox *Mailbox) DeleteEmails(emails []Email) error {
	imapClient, status, err := mbox.newClient()
	if err != nil {
		return err
	}

	defer imapClient.Logout()

	seqSet, err := uidSet(emails, status.UidValidity)
	if err != nil {
		return err
	}

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	err = imapClient.UidStore(seqSet, item, imap.DeletedFlag, nil)
	if err != nil {
		return err
	}

	return nil
}

func uidSet(emails []Email, uidValidity uint32) (*imap.SeqSet, error) {
	seqSet := new(imap.SeqSet)
	for _, message := range emails {
		if message.UIDValidity != uidValidity {
			return nil, ErrMailboxRecreated
		}
		seqSet.AddNum(message.UID)
	}
	return seqSet, nil
}

// GetUnread will find all unread emails in the folder and return them as a list.
func (mbox *Mailbox) GetUnread(markAsRead, delete bool) ([]Email, error) {
	imap.CharsetReader = charset.Reader
	var emails []Email

	imapClient, status, err := mbox.newClient()
	if err != nil {
		return emails, fmt.Errorf("failed to create IMAP connection: %s", err)
	}

	defer imapClient.Logout()

	// Search for unread emails
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	seqs, err := imapClient.UidSearch(criteria)
	if err != nil {
		return emails, err
	}

	if len(seqs) == 0 {
		return emails, nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(seqs...)
	section := &imap.BodySectionName{Peek: true, Partial: []int{0, maxRawMessageBytes + 1}}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, imap.FetchUid, section.FetchItem()}
	messages := make(chan *imap.Message)
	fetchErr := make(chan error, 1)
	processed := []Email{}
	go func() {
		fetchErr <- imapClient.UidFetch(seqset, items, messages)
	}()

	// Step through each email
	for msg := range messages {
		identity := Email{UID: msg.Uid, UIDValidity: status.UidValidity}
		processed = append(processed, identity)
		// Extract raw message body. I can't find a better way to do this with the emersion library
		var raw io.Reader
		for _, value := range msg.Body {
			raw = value
			break // There should only ever be one item in this map, but I'm not 100% sure
		}
		if raw == nil {
			continue
		}
		em, err := parseMessage(raw)
		if err != nil {
			log.Warn("Skipping unread message: ", err)
			continue
		}

		emtmp := Email{Email: em, UID: msg.Uid, UIDValidity: status.UidValidity}
		emails = append(emails, emtmp)
	}
	if err := <-fetchErr; err != nil {
		return emails, err
	}
	if markAsRead && len(processed) > 0 {
		seqSet, err := uidSet(processed, status.UidValidity)
		if err != nil {
			return emails, err
		}
		item := imap.FormatFlagsOp(imap.AddFlags, true)
		if err := imapClient.UidStore(seqSet, item, imap.SeenFlag, nil); err != nil {
			return emails, err
		}
	}
	return emails, nil
}

func parseMessage(raw io.Reader) (*email.Email, error) {
	buf, err := io.ReadAll(io.LimitReader(raw, maxRawMessageBytes+1))
	if err != nil {
		return nil, err
	}
	if len(buf) > maxRawMessageBytes {
		return nil, ErrMessageTooLarge
	}
	headerEnd := bytes.Index(buf, []byte("\r\n\r\n"))
	if headerEnd < 0 {
		headerEnd = bytes.Index(buf, []byte("\n\n"))
	}
	if headerEnd < 0 {
		headerEnd = len(buf)
	}
	if headerEnd > maxHeaderBytes {
		return nil, ErrMessageHeadersTooLarge
	}
	buf = bytes.ReplaceAll(buf, []byte("\r"), nil)
	return email.NewEmailFromReader(bytes.NewReader(buf))
}

// newClient will initiate a new IMAP connection with the given creds.
func (mbox *Mailbox) newClient() (*client.Client, *imap.MailboxStatus, error) {
	var imapClient *client.Client
	var err error
	restrictedDialer := dialer.Dialer()
	if mbox.TLS {
		config := new(tls.Config)
		config.InsecureSkipVerify = mbox.IgnoreCertErrors
		imapClient, err = client.DialWithDialerTLS(restrictedDialer, mbox.Host, config)
	} else {
		imapClient, err = client.DialWithDialer(restrictedDialer, mbox.Host)
	}
	if err != nil {
		return imapClient, nil, err
	}

	err = imapClient.Login(mbox.User, mbox.Pwd)
	if err != nil {
		return imapClient, nil, err
	}

	status, err := imapClient.Select(mbox.Folder, mbox.ReadOnly)
	if err != nil {
		return imapClient, nil, err
	}

	return imapClient, status, nil
}
