/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rest

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/mgmt/rest/rbody"
	"github.com/julienschmidt/httprouter"
)

var (
	tribeLogger = restLogger.WithFields(log.Fields{
		"_module": "rest-tribe",
	})

	ErrInvalidJSON           = errors.New("Invalid JSON")
	ErrAgreementDoesNotExist = errors.New("Agreement not found")
	ErrMemberNotFound        = errors.New("Member not found")
)

func (s *Server) getAgreements(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	res := &rbody.TribeListAgreement{}
	res.Agreements = s.tr.GetAgreements()
	respond(200, res, w)
}

func (s *Server) getAgreement(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tribeLogger = tribeLogger.WithField("_block", "getAgreement")
	name := p.ByName("name")
	if _, ok := s.tr.GetAgreements()[name]; !ok {
		fields := map[string]interface{}{
			"agreement_name": name,
		}
		tribeLogger.WithFields(fields).Error(ErrAgreementDoesNotExist)
		respond(400, rbody.FromSnapError(serror.New(ErrAgreementDoesNotExist, fields)), w)
		return
	}
	a := &rbody.TribeGetAgreement{}
	var serr serror.SnapError
	a.Agreement, serr = s.tr.GetAgreement(name)
	if serr != nil {
		tribeLogger.Error(serr)
		respond(400, rbody.FromSnapError(serr), w)
		return
	}
	respond(200, a, w)
}

func (s *Server) deleteAgreement(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tribeLogger = tribeLogger.WithField("_block", "deleteAgreement")
	name := p.ByName("name")
	if _, ok := s.tr.GetAgreements()[name]; !ok {
		fields := map[string]interface{}{
			"agreement_name": name,
		}
		tribeLogger.WithFields(fields).Error(ErrAgreementDoesNotExist)
		respond(400, rbody.FromSnapError(serror.New(ErrAgreementDoesNotExist, fields)), w)
		return
	}

	var serr serror.SnapError
	serr = s.tr.RemoveAgreement(name)
	if serr != nil {
		tribeLogger.Error(serr)
		respond(400, rbody.FromSnapError(serr), w)
		return
	}

	a := &rbody.TribeDeleteAgreement{}
	a.Agreements = s.tr.GetAgreements()
	respond(200, a, w)
}

func (s *Server) joinAgreement(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tribeLogger = tribeLogger.WithField("_block", "joinAgreement")
	name := p.ByName("name")
	if _, ok := s.tr.GetAgreements()[name]; !ok {
		fields := map[string]interface{}{
			"agreement_name": name,
		}
		tribeLogger.WithFields(fields).Error(ErrAgreementDoesNotExist)
		respond(400, rbody.FromSnapError(serror.New(ErrAgreementDoesNotExist, fields)), w)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		tribeLogger.Error(err)
		respond(500, rbody.FromError(err), w)
		return
	}

	m := struct {
		MemberName string `json:"member_name"`
	}{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		fields := map[string]interface{}{
			"error": err,
			"hint":  `The body of the request should be of the form '{"member_name": "some_value"}'`,
		}
		se := serror.New(ErrInvalidJSON, fields)
		tribeLogger.WithFields(fields).Error(ErrInvalidJSON)
		respond(400, rbody.FromSnapError(se), w)
		return
	}

	serr := s.tr.JoinAgreement(name, m.MemberName)
	if serr != nil {
		tribeLogger.Error(serr)
		respond(400, rbody.FromSnapError(serr), w)
		return
	}
	agreement, _ := s.tr.GetAgreement(name)
	respond(200, &rbody.TribeJoinAgreement{Agreement: agreement}, w)

}

func (s *Server) leaveAgreement(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tribeLogger = tribeLogger.WithField("_block", "leaveAgreement")
	name := p.ByName("name")
	if _, ok := s.tr.GetAgreements()[name]; !ok {
		fields := map[string]interface{}{
			"agreement_name": name,
		}
		tribeLogger.WithFields(fields).Error(ErrAgreementDoesNotExist)
		respond(400, rbody.FromSnapError(serror.New(ErrAgreementDoesNotExist, fields)), w)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		tribeLogger.Error(err)
		respond(500, rbody.FromError(err), w)
		return
	}

	m := struct {
		MemberName string `json:"member_name"`
	}{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		fields := map[string]interface{}{
			"error": err,
			"hint":  `The body of the request should be of the form '{"member_name": "some_value"}'`,
		}
		se := serror.New(ErrInvalidJSON, fields)
		tribeLogger.WithFields(fields).Error(ErrInvalidJSON)
		respond(400, rbody.FromSnapError(se), w)
		return
	}

	serr := s.tr.LeaveAgreement(name, m.MemberName)
	if serr != nil {
		tribeLogger.Error(serr)
		respond(400, rbody.FromSnapError(serr), w)
		return
	}
	agreement, _ := s.tr.GetAgreement(name)
	respond(200, &rbody.TribeLeaveAgreement{Agreement: agreement}, w)
}

func (s *Server) getMembers(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	members := s.tr.GetMembers()
	respond(200, &rbody.TribeMemberList{Members: members}, w)
}

func (s *Server) getMember(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tribeLogger = tribeLogger.WithField("_block", "getMember")
	name := p.ByName("name")
	member := s.tr.GetMember(name)
	if member == nil {
		fields := map[string]interface{}{
			"name": name,
		}
		tribeLogger.WithFields(fields).Error(ErrMemberNotFound)
		respond(404, rbody.FromSnapError(serror.New(ErrMemberNotFound, fields)), w)
		return
	}
	resp := &rbody.TribeMemberShow{
		Name: member.Name,
		Tags: member.Tags,
	}
	if member.PluginAgreement != nil {
		resp.PluginAgreement = member.PluginAgreement.Name
	}
	for k, t := range member.TaskAgreements {
		if len(t.Tasks) > 0 {
			resp.TaskAgreements = append(resp.TaskAgreements, k)
		}
	}
	respond(200, resp, w)
}

func (s *Server) addAgreement(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tribeLogger = tribeLogger.WithField("_block", "addAgreement")
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		tribeLogger.Error(err)
		respond(500, rbody.FromError(err), w)
		return
	}

	a := struct{ Name string }{}
	err = json.Unmarshal(b, &a)
	if err != nil {
		fields := map[string]interface{}{
			"error": err,
			"hint":  `The body of the request should be of the form '{"name": "agreement_name"}'`,
		}
		se := serror.New(ErrInvalidJSON, fields)
		tribeLogger.WithFields(fields).Error(ErrInvalidJSON)
		respond(400, rbody.FromSnapError(se), w)
		return
	}

	if a.Name == "" {
		fields := map[string]interface{}{
			"hint": `The body of the request should be of the form '{"name": "agreement_name"}'`,
		}
		se := serror.New(ErrInvalidJSON, fields)
		tribeLogger.WithFields(fields).Error(ErrInvalidJSON)
		respond(400, rbody.FromSnapError(se), w)
		return
	}

	err = s.tr.AddAgreement(a.Name)
	if err != nil {
		tribeLogger.WithField("agreement-name", a.Name).Error(err)
		respond(400, rbody.FromError(err), w)
		return
	}

	res := &rbody.TribeAddAgreement{}
	res.Agreements = s.tr.GetAgreements()

	respond(200, res, w)
}
