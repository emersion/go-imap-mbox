package mbox

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend/backendutil"
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

func setFlags(h message.Header, flags []string) {
	status := ""
	for _, flag := range flags {
		for s, f := range statuses {
			if f == flag {
				status += string(s)
				break
			}
		}
	}

	h.Set("X-Status", status)
}

func getMessage(e *message.Entity, msg *imap.Message, items []string) error {
	b, err := ioutil.ReadAll(e.Body)
	if err != nil {
		return err
	}

	br := bytes.NewReader(b)
	e.Body = br

	h := mail.Header{e.Header}

	for _, item := range items {
		switch item {
		case imap.BodyMsgAttr, imap.BodyStructureMsgAttr:
			extended := item == imap.BodyStructureMsgAttr
			msg.BodyStructure, _ = backendutil.FetchBodyStructure(e, extended)
			br.Seek(0, io.SeekStart)
		case imap.EnvelopeMsgAttr:
			msg.Envelope, _ = backendutil.FetchEnvelope(e.Header)
		case imap.FlagsMsgAttr:
			msg.Flags = getFlags(e.Header)
		case imap.InternalDateMsgAttr:
			msg.InternalDate, _ = h.Date()
		case imap.SizeMsgAttr:
			msg.Size = uint32(len(b)) // TODO: add headers size
		case imap.UidMsgAttr:
			// Nothing to do here
		default:
			section, err := imap.NewBodySectionName(item)
			if err != nil {
				break
			}

			msg.Body[section], _ = backendutil.FetchBodySection(e, section)
			br.Seek(0, io.SeekStart)
		}
	}

	return nil
}
