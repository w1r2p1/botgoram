package botgoram

import (
	"errors"
	"log"
	"regexp"

	"github.com/Patrolavia/botgoram/telegram"
)

var command_spliter *regexp.Regexp

func init() {
	cs, err := regexp.Compile(`^(\S+)(\s*.*)?$`)
	if err != nil {
		log.Fatalf("Cannot compile regular expression for spliting command, contact botgoram developer!")
	}
	command_spliter = cs
}

// ErrNoMatching denotes the message does not match the transitor
var ErrNoMatch error = errors.New("No matching transitor!")

const InitialState string = "" // predefined state id for initial state
// Transitor transits to next state according to message
//
// Which transitor to call
//
// Basically, we test message type to determine which transitor to call.
// If there is no transitor for this type, or transitor returns error code,
// which means match failed, the fallback transitors (if any) will be called.
//
// There is one exception: text messages. Text messages will be matched against
// special "Command" type before trying to call text transitors.
//
// Forwarded and replied message will go transitors handling forward/reply first,
// no matter which type it is. These messages will fallback to message type
// transitors when match failed.
//
// You should take care of not registering same transitor to a state twice, or
// the transitor will be called twice when match failed.
//
// Order of transitors
//
// Sometimes you need to register more than one transitor to a type. For example,
// an image bot might want to transit to different state according to image file
// format. The order we run transitor is just the order you register it.
type Transitor func(msg *telegram.Message, state State) (next string, err error)

// State describes how you can act with FSM and state data.
type State interface {
	Data() interface{}
	SetData(interface{})
	User() *telegram.User // who this state associate with
	Id() string           // retrive current state id
	Transit(id string)    // directly transit to another state without transitor
	// Transit again base on this state.
	// Retransit() have lower priority than Transit(id), if you call
	// Transit(id) anywhere before or after Retransit(), the state will
	// transit to id, without testing any transitor.
	Retransit()

	// register transitors by message types
	Register(mt telegram.MessageType, t Transitor)

	// Command is a special text message type, will be matched before text type.
	// A text message matches /^(\S+)(\s*.*)?$/ will go here before text type, and
	// we use first matching group to find out whcih transitor to call, case-sensitive.
	// (We use \S in regexp so you can define command in any language)
	RegisterCommand(cmd string, t Transitor)

	RegisterFallback(Transitor)
	test(msg *telegram.Message) (next string, err error)
	clone(user *telegram.User) State
	next() *string
	re() bool
}

type transitors []Transitor

func (ts transitors) test(msg *telegram.Message, cur State) (next string, err error) {
	err = ErrNoMatch
	for _, t := range ts {
		if next, err = t(msg, cur); err == nil {
			return
		}
	}
	return
}

type state struct {
	data      interface{}
	user      *telegram.User
	id        string
	forward   transitors
	reply     transitors
	types     map[telegram.MessageType]transitors
	command   map[string]transitors
	text      transitors
	fallback  transitors
	chain     *string
	retransit bool
}

func newState(id string) State {
	return &state{
		id:      id,
		types:   make(map[telegram.MessageType]transitors),
		command: make(map[string]transitors),
	}
}

func (s *state) clone(user *telegram.User) State {
	c := *s
	c.user = user
	return &c
}

func (s *state) Retransit() {
	s.retransit = true
}

func (s *state) re() bool {
	return s.retransit
}

func (s *state) Transit(id string) {
	s.chain = &id
}

func (s *state) next() *string {
	return s.chain
}

func (s *state) Data() interface{} {
	return s.data
}

func (s *state) SetData(data interface{}) {
	s.data = data
}

func (s *state) User() *telegram.User {
	return s.user
}

func (s *state) Id() string {
	return s.id
}

func (s *state) RegisterForward(t Transitor) {
	s.forward = append(s.forward, t)
}

func (s *state) RegisterReply(t Transitor) {
	s.reply = append(s.reply, t)
}

func (s *state) Register(mt telegram.MessageType, t Transitor) {
	s.types[mt] = append(s.types[mt], t)
}

func (s *state) RegisterCommand(cmd string, t Transitor) {
	s.command[cmd] = append(s.command[cmd], t)
}

func (s *state) RegisterFallback(t Transitor) {
	s.fallback = append(s.fallback, t)
}

func (s *state) test(msg *telegram.Message) (next string, err error) {
	do_test := func(ts transitors) (next string, err error) {
		if len(ts) == 0 {
			return next, ErrNoMatch
		}
		return ts.test(msg, s)
	}
	// TODO: find better way to test commands.
	test_cmd := func() (next string, err error) {
		err = ErrNoMatch
		matches := command_spliter.FindStringSubmatch(msg.Text)
		if len(matches) != 3 {
			return
		}
		cmd, ok := s.command[matches[1]]
		if !ok {
			return
		}
		return do_test(cmd)
	}

	// process forwarded message and replied message
	if msg.Forward != nil {
		if next, err = do_test(s.forward); err == nil {
			return
		}
	}
	if msg.ReplyTo != nil {
		if next, err = do_test(s.reply); err == nil {
			return
		}
	}

	msg_type := msg.Type()
	// process command message
	if msg_type == telegram.TEXT {
		if next, err = test_cmd(); err == nil {
			return
		}
	}
	if _, ok := s.types[msg_type]; ok {
		if next, err = do_test(s.types[msg_type]); err == nil {
			return
		}
	}
	next, err = do_test(s.fallback)
	return
}
