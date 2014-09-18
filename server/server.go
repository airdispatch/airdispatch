package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	adErrors "airdispat.ch/errors"
	"airdispat.ch/identity"
	"airdispat.ch/message"
	"airdispat.ch/routing"
	"airdispat.ch/wire"
)

// A structure that stores any errors generated by the Server framework
type ServerError struct {
	Location string
	Error    error
}

// This interface defines the functions that an
// Airdispatch server must respond to in order to properly function
type ServerDelegate interface {
	HandleError(err *ServerError)
	LogMessage(toLog ...string)

	SaveMessageDescription(desc *message.EncryptedMessage)

	RetrieveDataForUser(id string, author *identity.Address, forAddr *identity.Address) (*message.EncryptedMessage, io.ReadCloser)
	RetrieveMessageForUser(id string, author *identity.Address, forAddr *identity.Address) *message.EncryptedMessage
	RetrieveMessageListForUser(since uint64, author *identity.Address, forAddr *identity.Address) []*message.EncryptedMessage
}

// The server structure tahat holds all of the necessary instance variables
type Server struct {
	LocationName string
	Key          *identity.Identity
	Delegate     ServerDelegate
	Handlers     []Handler
	Router       routing.Router
	// Control Channels
	Start chan bool
	Quit  chan bool
}

// Function that starts the server on a specific port
func (s *Server) StartServer(port string) error {
	// Resolve the Address of the Server
	service := ":" + port
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", service)
	s.Delegate.LogMessage("Starting Server on " + service)

	// Start the Server
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	s.serverLoop(listener)
	return nil
}

// Sends an error to the Server Delegate
func (s *Server) handleError(location string, error error) {
	s.Delegate.HandleError(&ServerError{
		Location: location,
		Error:    error,
	})
}

// The loop that continues while waiting on clients to connect
func (s *Server) serverLoop(listener *net.TCPListener) {
	connections := make(chan net.Conn)

	// Loop forever, waiting for connections
	go func() {
		for {
			// Accept a Connection
			conn, err := listener.Accept()
			if err != nil {
				// Shutdown if the Error is Quit-Related
				select {
				case <-s.Quit:
					return
				default:
					s.handleError("Server Loop (Accepting New Client)", err)
					continue
				}
			}
			connections <- conn
		}
	}()

	if s.Start != nil {
		s.Start <- true
	}

	for {
		select {
		case conn := <-connections:
			// Concurrently handle the connection
			go s.handleClient(conn)
		case <-s.Quit:
			// Close the listener
			close(s.Quit)
			listener.Close()
			return
		}
	}
}

// Called when a client connects
func (s *Server) handleClient(conn net.Conn) {
	s.Delegate.LogMessage("Serving", conn.RemoteAddr().String())
	tNow := time.Now()
	defer s.Delegate.LogMessage("Finished with", conn.RemoteAddr().String(), "in", time.Since(tNow).String())

	// Close the Connection after Handling
	defer conn.Close()

	// Read in the Message
	newMessage, err := message.ReadMessageFromConnection(conn)
	if err != nil {
		// There is nothing we can do if we can't read the message.
		s.handleError("Read Message From Connection", err)
		adErrors.CreateError(adErrors.UnexpectedError, "Unable to read message properly.", s.Key.Address).Send(s.Key, conn)
		return
	}

	_, ok := newMessage.Header[s.Key.Address.String()]
	if ok {
		signedMessage, err := newMessage.Decrypt(s.Key)
		if err != nil {
			s.handleError("Decrypt Message", err)
			adErrors.CreateError(adErrors.UnexpectedError, "Unable to decrypt message.", s.Key.Address).Send(s.Key, conn)
			return
		}

		if !signedMessage.Verify() {
			s.handleError("Verify Signature", errors.New("Unable to Verify Signature on Message"))
			adErrors.CreateError(adErrors.InvalidSignature, "Message contains invalid signature.", s.Key.Address).Send(s.Key, conn)
			return
		}

		data, mesType, h, err := signedMessage.ReconstructMessageWithTimestamp()

		if err != nil {
			s.handleError("Verifying Message Structure", err)
			adErrors.CreateError(adErrors.UnexpectedError, "Unable to unpack transfer message.", s.Key.Address).Send(s.Key, conn)
			return
		}

		// Switch based on the Message Type
		switch mesType {
		case wire.TransferMessageCode:
			s.handleTransferMessage(data, h, conn)
		case wire.TransferMessageListCode:
			s.handleTransferMessageList(data, h, conn)
		}

		returnAddress := h.From
		// Lookup from Router if Return Address is not Sendable
		if !h.From.CanSend() {
			if s.Router == nil {
				adErrors.CreateError(adErrors.UnexpectedError, "No router to lookup your address. Must provide return information.", s.Key.Address).Send(s.Key, conn)
				return
			}
			if h.From.Alias != "" {
				// Lookup by Alias
				returnAddress, err = s.Router.LookupAlias(h.From.Alias, routing.LookupTypeDEFAULT)
				if err != nil {
					s.handleError("Looking up Return Address", err)
					adErrors.CreateError(adErrors.UnexpectedError, "Cannot lookup return address.", s.Key.Address).Send(s.Key, conn)
					return
				}
			} else {
				// Lookup by Address
				returnAddress, err = s.Router.Lookup(h.From.String(), routing.LookupTypeDEFAULT)
				if err != nil {
					s.handleError("Looking up Return Address", err)
					adErrors.CreateError(adErrors.UnexpectedError, "Cannot lookup return address.", s.Key.Address).Send(s.Key, conn)
					return
				}
			}
		}

		// Attempt Sub Handlers for Extra Message Types
		for _, v := range s.Handlers {
			if v.HandlesType(mesType) {
				response, err := v.HandleMessage(mesType, data, h)
				if err != nil {
					s.handleError("Sub-handler", err)
				}

				if len(response) == 0 {
					adErrors.CreateError(adErrors.UnexpectedError, "No response from handler.", s.Key.Address).Send(s.Key, conn)
				}

				for _, v := range response {
					err := message.SignAndSendToConnection(v, s.Key, returnAddress, conn)
					if err != nil {
						fmt.Println("Got error sending message from Handler: ", err)
					}
				}

				return
			}
		}
		adErrors.CreateError(adErrors.UnexpectedError, "Unable to handle message type.", s.Key.Address).Send(s.Key, conn)
	} else {
		s.handleMessageDescription(newMessage)
	}
}

// Send the Message to the Delegate
func (s *Server) handleMessageDescription(desc *message.EncryptedMessage) {
	s.Delegate.SaveMessageDescription(desc)
}

// Function that Handles a DataRetrieval Message
func (s *Server) handleTransferMessage(desc []byte, h message.Header, conn net.Conn) {
	txMessage, err := CreateTransferMessageFromBytes(desc, h)
	if err != nil {
		adErrors.CreateError(adErrors.UnexpectedError, "Unable to unpack transfer message.", s.Key.Address).Send(s.Key, conn)
		return
	}

	var mail *message.EncryptedMessage
	var reader io.ReadCloser

	if txMessage.Data {
		mail, reader = s.Delegate.RetrieveDataForUser(txMessage.Name, txMessage.Author, txMessage.h.From)
	} else {
		mail = s.Delegate.RetrieveMessageForUser(txMessage.Name, txMessage.Author, txMessage.h.From)
	}

	// If mail is nil, then there is no message.
	if mail == nil {
		s.handleError("Loading message from Server", errors.New("Couldn't find message named"+txMessage.Name))
		adErrors.CreateError(adErrors.MessageNotFound, "That message doesn't exist.", s.Key.Address).Send(s.Key, conn)
		return
	}

	err = mail.SendMessageToConnection(conn)
	if err != nil {
		s.handleError("Sign and Send Mail", err)
		adErrors.CreateError(adErrors.InternalError, "Unable to pack return message.", s.Key.Address).Send(s.Key, conn)
		return
	}

	if txMessage.Data {
		io.Copy(conn, reader)
		reader.Close()
	}
}

func (s *Server) handleTransferMessageList(desc []byte, h message.Header, conn net.Conn) {
	txMessage, err := CreateTransferMessageListFromBytes(desc, h)
	if err != nil {
		adErrors.CreateError(adErrors.UnexpectedError, "Unable to unpack transfer message list.", s.Key.Address).Send(s.Key, conn)
		return
	}

	mail := s.Delegate.RetrieveMessageListForUser(txMessage.Since, txMessage.Author, txMessage.h.From)
	if mail == nil {
		s.handleError("Loading message from Server", errors.New("Couldn't find message"))
		adErrors.CreateError(adErrors.MessageNotFound, "Couldn't find any messages for that user.", s.Key.Address).Send(s.Key, conn)
		return
	}

	ml := &MessageList{
		Length: uint64(len(mail)),
		h:      message.CreateHeader(s.Key.Address, txMessage.h.From),
	}

	err = message.SignAndSendToConnection(ml, s.Key, txMessage.h.From, conn)
	if err != nil {
		s.handleError("Sending message list to connection.", err)
		adErrors.CreateError(adErrors.InternalError, "Unable to pack return message.", s.Key.Address).Send(s.Key, conn)
		return
	}

	for _, v := range mail {
		err := v.SendMessageToConnection(conn)
		if err != nil {
			s.handleError("Sending public message to connection.", err)
		}
	}
}
