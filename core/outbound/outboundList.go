package outbound

import (
	"github.com/holyshawn/overture/core/config"
	"github.com/miekg/dns"
)

type OutboundListType struct {
	OutboundList []*Outbound

	ResponseMessage *dns.Msg
	QuestionMessage *dns.Msg

	DNSUpstream []*config.DNSUpstream
	InboundIP   string
}

func (ol *OutboundListType) GetResponse(IsCache bool) {

	ch := make(chan *dns.Msg, len(ol.OutboundList))

	go func() {
		for _, o := range ol.OutboundList {
			o.ExchangeFromRemote(IsCache)
			ch <- o.ResponseMessage
		}
	}()

	for i := 0; i < len(ol.OutboundList); i++ {
		if m := <-ch; m != nil {
			ol.ResponseMessage = m
			return
		}
	}
}

func (ol *OutboundListType) SetOutboundList(q *dns.Msg, inboundIP string, dl []*config.DNSUpstream) {

	ol.QuestionMessage = q
	ol.InboundIP = inboundIP
	ol.DNSUpstream = dl

	for _, u := range dl {

		o := NewOutbound(q, u, inboundIP)
		ol.OutboundList = append(ol.OutboundList, o)
	}
}

func (ol *OutboundListType) UpdateDNSUpstream(dl []*config.DNSUpstream) {

	for _, u := range dl {
		for _, o := range ol.OutboundList {
			o.DNSUpstream = u
		}
	}
}

func (ol *OutboundListType) EqualDNSUpstream(dl []*config.DNSUpstream) bool {

	for _, u := range dl {
		for _, o := range ol.OutboundList {
			if o.DNSUpstream != u {
				return false
			}
		}
	}

	return true
}

func NewOutboundList(q *dns.Msg, dl []*config.DNSUpstream, inboundIP string) *OutboundListType {

	ol := new(OutboundListType)

	ol.SetOutboundList(q, inboundIP, dl)

	return ol
}
