package mbox

import (
	"io"
	"os"
	"time"

	"github.com/blabber/mbox"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message"
)

type mailbox struct {
	info *imap.MailboxInfo
	f *os.File
	readOnly bool
	subscribed bool
}

func (m *mailbox) Name() string {
	return m.info.Name
}

func (m *mailbox) Info() (*imap.MailboxInfo, error) {
	return m.info, nil
}

func (m *mailbox) scanner() (*mbox.Scanner, error) {
	if _, err := m.f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return mbox.NewScanner(m.f), nil
}

func (m *mailbox) Status(items []string) (*imap.MailboxStatus, error) {
	status := imap.NewMailboxStatus(m.info.Name, items)
	status.ReadOnly = m.readOnly
	status.PermanentFlags = []string{imap.SeenFlag, imap.AnsweredFlag, imap.FlaggedFlag, imap.DeletedFlag, imap.DraftFlag}

	s, err := m.scanner()
	if err != nil {
		return nil, err
	}
	for s.Next() {
		msg := s.Message()

		status.Messages++

		if uid, err := getUID(message.Header(msg.Header)); err == nil && uid > status.UidNext {
			status.UidNext = uid
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	status.UidNext++
	status.UidValidity = 1
	return status, nil
}

func (m *mailbox) Subscribe() error {
	m.subscribed = true
	return nil
}

func (m *mailbox) Unsubscribe() error {
	m.subscribed = false
	return nil
}

func (m *mailbox) Check() error {
	return m.f.Sync()
}

func (m *mailbox) ListMessages(isUID bool, seqset *imap.SeqSet, items []string, ch chan<- *imap.Message) error {
	defer close(ch)

	s, err := m.scanner()
	if err != nil {
		return err
	}

	var seqnum uint32
	for s.Next() {
		msg := s.Message()
		seqnum++

		uid, err := getUID(message.Header(msg.Header))
		if err != nil {
			continue
		}

		// Filter messages with seqset
		if (isUID && !seqset.Contains(uid)) || (!isUID && !seqset.Contains(seqnum)) {
			continue
		}

		e := message.NewEntity(message.Header(msg.Header), msg.Body)

		imapMsg := imap.NewMessage(seqnum, items)
		imapMsg.Uid = uid
		if err := getMessage(e, imapMsg, items); err != nil {
			return err
		}
		ch <- imapMsg
	}

	return s.Err()
}

func (m *mailbox) SearchMessages(uid bool, criteria *imap.SearchCriteria) ([]uint32, error) {
	return nil, nil // TODO
}

func (m *mailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {
	return nil // TODO
}

func (m *mailbox) UpdateMessagesFlags(uid bool, seqset *imap.SeqSet, operation imap.FlagsOp, flags []string) error {
	return nil // TODO
}

func (m *mailbox) CopyMessages(uid bool, seqset *imap.SeqSet, dest string) error {
	return nil // TODO
}

func (m *mailbox) Expunge() error {
	return nil // TODO
}
