package mbox

import (
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
)

func getUID(h message.Header) (uint32, error) {
	uid, err := strconv.ParseUint(h.Get("X-UID"), 10, 32)
	return uint32(uid), err
}

var statuses = map[rune]string{
	'A': imap.AnsweredFlag,
	'F': imap.FlaggedFlag,
	'T': imap.DraftFlag,
	'D': imap.DeletedFlag,
	'R': imap.SeenFlag,
}

func getFlags(h message.Header) []string {
	var flags []string
	for _, c := range strings.ToUpper(h.Get("X-Status")) {
		if f, ok := statuses[c]; ok {
			flags = append(flags, f)
		}
	}
	return flags
}

func getMessage(e *message.Entity, msg *imap.Message, items []string) error {
	b, err := ioutil.ReadAll(e.Body)
	if err != nil {
		return err
	}

	h := mail.Header{e.Header}

	for _, item := range items {
		switch item {
		case imap.BodyMsgAttr, imap.BodyStructureMsgAttr:
			// TODO
		case imap.EnvelopeMsgAttr:
			e := &imap.Envelope{
				InReplyTo: h.Get("In-Reply-To"),
				MessageId: h.Get("Message-Id"),
				// TODO
			}
			e.Date, _ = h.Date()
			msg.Envelope = e
		case imap.FlagsMsgAttr:
			msg.Flags = getFlags(e.Header)
		case imap.InternalDateMsgAttr:
			msg.InternalDate, _ = h.Date()
		case imap.SizeMsgAttr:
			msg.Size = uint32(len(b)) // TODO: add headers
		case imap.UidMsgAttr:
			// Nothing to do here
		default:
			section, err := imap.NewBodySectionName(item)
			if err != nil {
				break
			}

			// TODO
			_ = section
		}
	}

	return nil
}
