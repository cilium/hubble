// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package policy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/bits"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"

	cilium "github.com/cilium/proxy/go/cilium/api"
	"github.com/cilium/proxy/pkg/policy/api/kafka"

	"github.com/sirupsen/logrus"
	k8sTypes "k8s.io/apimachinery/pkg/types"

	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/container/bitlpm"
	"github.com/cilium/cilium/pkg/container/versioned"
	"github.com/cilium/cilium/pkg/iana"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/labels"
	"github.com/cilium/cilium/pkg/lock"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/option"
	"github.com/cilium/cilium/pkg/policy/api"
	"github.com/cilium/cilium/pkg/policy/trafficdirection"
	"github.com/cilium/cilium/pkg/policy/types"
	"github.com/cilium/cilium/pkg/u8proto"
)

type AuthType = types.AuthType
type AuthTypes = types.AuthTypes
type AuthRequirement = types.AuthRequirement

// authmap maps remote selectors to their needed AuthTypes, if any
type authMap map[CachedSelector]types.AuthTypes

// covers returns true if 'l4rule' has the effect needed for the 'l3l4rule', when 'l4rule' is added
// to the datapath, due to the l4-only rule matching if l3l4-rule is not present. This determination
// can be done here only when both rules have the same port number (or both have a wildcarded port).
func (l4rule *PerSelectorPolicy) covers(l3l4rule *PerSelectorPolicy) bool {
	// Deny takes highest precedence so it is dealt with first
	if l4rule != nil && l4rule.IsDeny {
		// l4-only deny takes precedence
		return true
	} else if l3l4rule != nil && l3l4rule.IsDeny {
		// Must not skip if l3l4 rule is deny while l4-only rule is not
		return false
	}

	// Can not skip if rules have different auth types.  In all other cases the auth type from
	// the wildcardRule can be used also for the current rule.
	// Note that the caller must deal with inheriting redirect from wildcardRule to currentRule,
	// if any.
	if l3l4rule.getAuthRequirement() != l4rule.getAuthRequirement() {
		return false
	}

	l3l4IsRedirect := l3l4rule.IsRedirect()
	l4OnlyIsRedirect := l4rule.IsRedirect()
	if l3l4IsRedirect && !l4OnlyIsRedirect {
		// Can not skip if l3l4-rule is redirect while l4-only is not
		return false
	} else if l3l4IsRedirect && l4OnlyIsRedirect &&
		(l3l4rule.Listener != l4rule.Listener || l3l4rule.Priority != l4rule.Priority) {
		// L3l4 rule has a different listener or priority, it can not be skipped
		return false
	}

	// else can skip
	return true
}

// TLS context holds the secret values resolved from an 'api.TLSContext'
type TLSContext struct {
	TrustedCA        string `json:"trustedCA,omitempty"`
	CertificateChain string `json:"certificateChain,omitempty"`
	PrivateKey       string `json:"privateKey,omitempty"`
	// Secret holds the name of the Secret that was referenced in the Policy
	Secret k8sTypes.NamespacedName
	// FromFile is true if the values in the keys above were read from the filesystem
	// and not a Kubernetes Secret
	FromFile bool
}

// Equal returns true if 'a' and 'b' have the same contents.
func (a *TLSContext) Equal(b *TLSContext) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

// MarshalJSON marsahls a redacted version of the TLSContext. We want
// to see which fields are present, but not reveal their values in any
// logs, etc.
func (t *TLSContext) MarshalJSON() ([]byte, error) {
	type tlsContext TLSContext
	var redacted tlsContext
	if t.TrustedCA != "" {
		redacted.TrustedCA = "[redacted]"
	}
	if t.CertificateChain != "" {
		redacted.CertificateChain = "[redacted]"
	}
	if t.PrivateKey != "" {
		redacted.PrivateKey = "[redacted]"
	}
	return json.Marshal(&redacted)
}

type StringSet map[string]struct{}

func (a StringSet) Equal(b StringSet) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, exists := b[k]; !exists {
			return false
		}
	}
	return true
}

// NewStringSet returns a StringSet initialized from slice of strings.
// Returns nil for an empty slice
func NewStringSet(from []string) StringSet {
	if len(from) == 0 {
		return nil
	}
	set := make(StringSet, len(from))
	for _, s := range from {
		set[s] = struct{}{}
	}
	return set
}

// Merge returns StringSet with strings from both a and b.
// Returns a or b, possibly with modifications.
func (a StringSet) Merge(b StringSet) StringSet {
	if len(a) == 0 {
		return b
	}
	for s := range b {
		a[s] = struct{}{}
	}
	return a
}

// PerSelectorPolicy contains policy rules for a CachedSelector, i.e. for a
// selection of numerical identities.
type PerSelectorPolicy struct {
	// TerminatingTLS is the TLS context for the connection terminated by
	// the L7 proxy.  For egress policy this specifies the server-side TLS
	// parameters to be applied on the connections originated from the local
	// POD and terminated by the L7 proxy. For ingress policy this specifies
	// the server-side TLS parameters to be applied on the connections
	// originated from a remote source and terminated by the L7 proxy.
	TerminatingTLS *TLSContext `json:"terminatingTLS,omitempty"`

	// OriginatingTLS is the TLS context for the connections originated by
	// the L7 proxy.  For egress policy this specifies the client-side TLS
	// parameters for the upstream connection originating from the L7 proxy
	// to the remote destination. For ingress policy this specifies the
	// client-side TLS parameters for the connection from the L7 proxy to
	// the local POD.
	OriginatingTLS *TLSContext `json:"originatingTLS,omitempty"`

	// ServerNames is a list of allowed TLS SNI values. If not empty, then
	// TLS must be present and one of the provided SNIs must be indicated in the
	// TLS handshake.
	ServerNames StringSet `json:"serverNames,omitempty"`

	// isRedirect is 'true' when traffic must be redirected
	isRedirect bool `json:"-"`

	// Listener is an optional fully qualified name of a Envoy Listner defined in a CiliumEnvoyConfig CRD that should be
	// used for this traffic instead of the default listener
	Listener string `json:"listener,omitempty"`

	// Priority of the listener used when multiple listeners would apply to the same
	// MapStateEntry.
	// Lower numbers indicate higher priority. Except for the default 0, which indicates the
	// lowest priority.  If higher priority desired, a low unique number like 1, 2, or 3 should
	// be explicitly specified here.
	Priority uint8 `json:"priority,omitempty"`

	// Pre-computed HTTP rules, computed after rule merging is complete
	EnvoyHTTPRules *cilium.HttpNetworkPolicyRules `json:"-"`

	// CanShortCircuit is true if all 'EnvoyHTTPRules' may be
	// short-circuited by other matches.
	CanShortCircuit bool `json:"-"`

	api.L7Rules

	// Authentication is the kind of cryptographic authentication required for the traffic to be allowed
	// at L3, if any.
	Authentication *api.Authentication `json:"auth,omitempty"`

	// IsDeny is set if this L4Filter contains should be denied
	IsDeny bool `json:",omitempty"`
}

// Equal returns true if 'a' and 'b' represent the same L7 Rules
func (a *PerSelectorPolicy) Equal(b *PerSelectorPolicy) bool {
	return a == nil && b == nil || a != nil && b != nil &&
		a.TerminatingTLS.Equal(b.TerminatingTLS) &&
		a.OriginatingTLS.Equal(b.OriginatingTLS) &&
		a.ServerNames.Equal(b.ServerNames) &&
		a.isRedirect == b.isRedirect &&
		a.Listener == b.Listener &&
		a.Priority == b.Priority &&
		(a.Authentication == nil && b.Authentication == nil || a.Authentication != nil && a.Authentication.DeepEqual(b.Authentication)) &&
		a.IsDeny == b.IsDeny &&
		a.L7Rules.DeepEqual(&b.L7Rules)
}

// GetListener returns the listener of the PerSelectorPolicy.
func (a *PerSelectorPolicy) GetListener() string {
	if a == nil {
		return ""
	}
	return a.Listener
}

// GetPriority returns the pritority of the listener of the PerSelectorPolicy.
func (a *PerSelectorPolicy) GetPriority() uint8 {
	if a == nil {
		return 0
	}
	return a.Priority
}

// getAuthType returns AuthType for the api.Authentication
func getAuthType(auth *api.Authentication) (bool, AuthType) {
	if auth == nil {
		return false, types.AuthTypeDisabled
	}
	switch auth.Mode {
	case api.AuthenticationModeDisabled:
		return true, types.AuthTypeDisabled
	case api.AuthenticationModeRequired:
		return true, types.AuthTypeSpire
	case api.AuthenticationModeAlwaysFail:
		return true, types.AuthTypeAlwaysFail
	default:
		return false, types.AuthTypeDisabled
	}
}

// getAuthType returns the AuthType of the L4Filter.
func (a *PerSelectorPolicy) getAuthType() (bool, AuthType) {
	if a == nil {
		return false, types.AuthTypeDisabled
	}
	return getAuthType(a.Authentication)
}

// GetAuthRequirement returns the AuthRequirement of the L4Filter.
func (a *PerSelectorPolicy) getAuthRequirement() AuthRequirement {
	if a == nil {
		return AuthRequirement(types.AuthTypeDisabled)
	}
	explicit, authType := getAuthType(a.Authentication)
	req := AuthRequirement(authType)
	if explicit {
		req |= types.AuthTypeIsExplicit
	}
	return req
}

// IsRedirect returns true if the L7Rules are a redirect.
func (a *PerSelectorPolicy) IsRedirect() bool {
	return a != nil && a.isRedirect
}

// HasL7Rules returns whether the `L7Rules` contains any L7 rules.
func (a *PerSelectorPolicy) HasL7Rules() bool {
	return !a.L7Rules.IsEmpty()
}

// L7DataMap contains a map of L7 rules per endpoint where key is a CachedSelector
type L7DataMap map[CachedSelector]*PerSelectorPolicy

func (l7 L7DataMap) MarshalJSON() ([]byte, error) {
	if len(l7) == 0 {
		return []byte("[]"), nil
	}

	/* First, create a sorted slice of the selectors so we can get
	 * consistent JSON output */
	selectors := make(types.CachedSelectorSlice, 0, len(l7))
	for cs := range l7 {
		selectors = append(selectors, cs)
	}
	sort.Sort(selectors)

	/* Now we can iterate the slice and generate JSON entries. */
	var err error
	buffer := bytes.NewBufferString("[")
	for _, cs := range selectors {
		buffer.WriteString("{\"")
		buffer.WriteString(cs.String())
		buffer.WriteString("\":")
		b, err := json.Marshal(l7[cs])
		if err == nil {
			buffer.Write(b)
		} else {
			buffer.WriteString("\"L7DataMap error: ")
			buffer.WriteString(err.Error())
			buffer.WriteString("\"")
		}
		buffer.WriteString("},")
	}
	buffer.Truncate(buffer.Len() - 1) // Drop the final ","
	buffer.WriteString("]")

	return buffer.Bytes(), err
}

// ShallowCopy returns a shallow copy of the L7DataMap.
func (l7 L7DataMap) ShallowCopy() L7DataMap {
	m := make(L7DataMap, len(l7))
	for k, v := range l7 {
		m[k] = v
	}
	return m
}

// L7ParserType is the type used to indicate what L7 parser to use.
// Consts are defined for all well known L7 parsers.
// Unknown string values are created for key-value pair policies, which
// are then transparently used in redirect configuration.
type L7ParserType string

func (l7 L7ParserType) String() string {
	return (string)(l7)
}

const (
	// ParserTypeNone represents the case where no parser type is provided.
	ParserTypeNone L7ParserType = ""
	// ParserTypeTLS is used for TLS origination, termination, or SNI filtering without any L7
	// parsing. If TLS policies are used with HTTP rules, ParserTypeHTTP is used instead.
	ParserTypeTLS L7ParserType = "tls"
	// ParserTypeCRD is used with a custom CiliumEnvoyConfig redirection. Incompatible with any
	// parser type with L7 enforcement (HTTP, Kafka, proxylib), as the custom Listener generally
	// does not support them.
	ParserTypeCRD L7ParserType = "crd"
	// ParserTypeHTTP specifies a HTTP parser type
	ParserTypeHTTP L7ParserType = "http"
	// ParserTypeKafka specifies a Kafka parser type
	ParserTypeKafka L7ParserType = "kafka"
	// ParserTypeDNS specifies a DNS parser type
	ParserTypeDNS L7ParserType = "dns"
)

// redirectTypes is a bitmask of redirection types of multiple filters
type redirectTypes uint16

const (
	// redirectTypeDNS bit is set when policy contains a redirection to DNS proxy
	redirectTypeDNS redirectTypes = 1 << iota
	// redirectTypeEnvoy bit is set when policy contains a redirection to Envoy
	redirectTypeEnvoy
	// redirectTypeProxylib bits are set when policy contains a redirection to Proxylib (via
	// Envoy)
	redirectTypeProxylib redirectTypes = 1<<iota | redirectTypeEnvoy

	// redirectTypeNone represents the case where there is no proxy redirect
	redirectTypeNone redirectTypes = redirectTypes(0)
)

func (from L7ParserType) canPromoteTo(to L7ParserType) bool {
	switch from {
	case ParserTypeNone:
		// ParserTypeNone can be promoted to any other type
		return true
	case ParserTypeTLS:
		// ParserTypeTLS can be promoted to any other type, except for DNS or CRD,
		// but ParserTypeTLS can not be demoted to ParserTypeNone
		if to != ParserTypeNone && to != ParserTypeDNS && to != ParserTypeCRD {
			return true
		}
	}
	return false
}

// Merge ParserTypes 'a' to 'b' if possible
func (a L7ParserType) Merge(b L7ParserType) (L7ParserType, error) {
	if a == b {
		return a, nil
	}
	if a.canPromoteTo(b) {
		return b, nil
	}
	if b.canPromoteTo(a) {
		return a, nil
	}
	return ParserTypeNone, fmt.Errorf("cannot merge conflicting L7 parsers (%s/%s)", a, b)
}

// hasWildcard checks if the L7Rules contains a wildcard rule for the given parser type.
func hasWildcard(rules *api.L7Rules, parserType L7ParserType) bool {
	if rules == nil {
		return false
	}

	switch {
	case parserType == ParserTypeDNS:
		for _, rule := range rules.DNS {
			if rule.MatchPattern == "*" {
				return true
			}
		}
	case parserType == ParserTypeHTTP:
		for _, rule := range rules.HTTP {
			if rule.Path == "" && rule.Method == "" && rule.Host == "" &&
				len(rule.Headers) == 0 && len(rule.HeaderMatches) == 0 {
				return true
			}
		}
	case parserType == ParserTypeKafka:
		for _, rule := range rules.Kafka {
			if rule.Topic == "" {
				return true
			}
		}
	case rules.L7Proto != "":
		// For custom L7 rules
		for _, rule := range rules.L7 {
			if len(rule) == 0 {
				return true
			}
		}
	default:
		// Unsupported parser type
	}

	return false
}

// addWildcard adds a wildcard rule to the L7Rules for the given parser type.
// It returns a copy of the rules with the wildcard rule added.
func addWildcard(rules *api.L7Rules, parserType L7ParserType) *api.L7Rules {
	rulesCopy := *rules
	result := &rulesCopy

	switch {
	case parserType == ParserTypeDNS:
		if len(rules.DNS) > 0 {
			result.DNS = append(result.DNS, api.PortRuleDNS{MatchPattern: "*"})
		}
	case parserType == ParserTypeHTTP:
		if len(rules.HTTP) > 0 {
			result.HTTP = append(result.HTTP, api.PortRuleHTTP{})
		}
	case parserType == ParserTypeKafka:
		if len(rules.Kafka) > 0 {
			result.Kafka = append(result.Kafka, kafka.PortRule{})
		}
	case rules.L7Proto != "":
		// For custom L7 rules with L7Proto
		if len(rules.L7) > 0 {
			result.L7 = append(result.L7, api.PortRuleL7{})
		}
	default:
		// Unsupported parser type
	}

	return result
}

// ensureWildcard ensures that the L7Rules contains a wildcard rule for the given parser type.
// It returns a copy of the rules with the wildcard rule added if it wasn't already present.
func ensureWildcard(rules *api.L7Rules, parserType L7ParserType) *api.L7Rules {
	if rules == nil {
		return nil
	}

	if hasWildcard(rules, parserType) {
		return rules
	}

	return addWildcard(rules, parserType)
}

// L4Filter represents the policy (allowed remote sources / destinations of
// traffic) that applies at a specific L4 port/protocol combination (including
// all ports and protocols), at either ingress or egress. The policy here is
// specified in terms of selectors that are mapped to security identities via
// the selector cache.
type L4Filter struct {
	// Port is the destination port to allow. Port 0 indicates that all traffic
	// is allowed at L4.
	Port uint16 `json:"port"`
	// EndPort is zero for a singular port
	EndPort  uint16 `json:"endPort,omitempty"`
	PortName string `json:"port-name,omitempty"`
	// Protocol is the L4 protocol to allow or NONE
	Protocol api.L4Proto `json:"protocol"`
	// U8Proto is the Protocol in numeric format, or 0 for NONE
	U8Proto u8proto.U8proto `json:"-"`
	// wildcard is the cached selector representing a wildcard in this filter, if any.
	// This is nil the wildcard selector in not in 'PerSelectorPolicies'.
	// When the wildcard selector is in 'PerSelectorPolicies' this is set to that
	// same selector, which can then be used as a map key to find the corresponding
	// L4-only L7 policy (which can be nil).
	wildcard CachedSelector
	// PerSelectorPolicies is a map of policies for selectors, including any L7 rules passed to
	// the L7 proxy. nil values represent cached selectors that have selector-specific policy
	// restriction (such as no L7 rules). Holds references to the cached selectors, which must
	// be released!
	PerSelectorPolicies L7DataMap `json:"l7-rules,omitempty"`
	// L7Parser specifies the L7 protocol parser (optional). If specified as
	// an empty string, then means that no L7 proxy redirect is performed.
	L7Parser L7ParserType `json:"-"`
	// Ingress is true if filter applies at ingress; false if it applies at egress.
	Ingress bool `json:"-"`
	// RuleOrigin tracks which policy rules (identified by labels) are the origin for this L3/L4
	// (i.e. selector and port) filter. This information is used when distilling a policy to an
	// EndpointPolicy, to track which policy rules were involved for a specific verdict.
	// Each LabelArrayList is in sorted order.
	RuleOrigin map[CachedSelector]labels.LabelArrayList `json:"-"`

	// This reference is circular, but it is cleaned up at Detach()
	policy atomic.Pointer[L4Policy]
}

// SelectsAllEndpoints returns whether the L4Filter selects all
// endpoints, which is true if the wildcard endpoint selector is present in the
// map.
func (l4 *L4Filter) SelectsAllEndpoints() bool {
	for cs := range l4.PerSelectorPolicies {
		if cs.IsWildcard() {
			return true
		}
	}
	return false
}

// CopyL7RulesPerEndpoint returns a shallow copy of the PerSelectorPolicies of the
// L4Filter.
func (l4 *L4Filter) CopyL7RulesPerEndpoint() L7DataMap {
	return l4.PerSelectorPolicies.ShallowCopy()
}

// GetL7Parser returns the L7ParserType of the L4Filter.
func (l4 *L4Filter) GetL7Parser() L7ParserType {
	return l4.L7Parser
}

// GetIngress returns whether the L4Filter applies at ingress or egress.
func (l4 *L4Filter) GetIngress() bool {
	return l4.Ingress
}

// GetPort returns the port at which the L4Filter applies as a uint16.
func (l4 *L4Filter) GetPort() uint16 {
	return l4.Port
}

// Equals returns true if two L4Filters are equal
func (l4 *L4Filter) Equals(bL4 *L4Filter) bool {
	if l4.Port == bL4.Port &&
		l4.EndPort == bL4.EndPort &&
		l4.PortName == bL4.PortName &&
		l4.Protocol == bL4.Protocol &&
		l4.Ingress == bL4.Ingress &&
		l4.L7Parser == bL4.L7Parser &&
		l4.wildcard == bL4.wildcard {

		if len(l4.PerSelectorPolicies) != len(bL4.PerSelectorPolicies) {
			return false
		}
		for k, v := range l4.PerSelectorPolicies {
			bV, ok := bL4.PerSelectorPolicies[k]
			if !ok || !bV.Equal(v) {
				return false
			}

		}
		return true
	}
	return false
}

// ChangeState allows caller to revert changes made by (multiple) toMapState call(s)
// All fields are maps so we can pass this by value.
type ChangeState struct {
	Adds    Keys        // Added or modified keys, if not nil
	Deletes Keys        // deleted keys, if not nil
	old     mapStateMap // Old values of all modified or deleted keys, if not nil
}

// NewRevertState returns an empty ChangeState suitable for reverting MapState changes.
// The private 'old' field is initialized so that old state can be restored if need be.
func NewRevertState() ChangeState {
	return ChangeState{
		Adds: make(Keys),
		old:  make(mapStateMap),
	}
}

func (c *ChangeState) Empty() bool {
	return len(c.Adds)+len(c.Deletes)+len(c.old) == 0
}

// Size returns the total number of Adds minus
// the total number of true Deletes (Deletes
// that are not also in Adds). The return value
// can be negative.
func (c *ChangeState) Size() int {
	deleteLen := 0
	for k := range c.Deletes {
		if _, ok := c.Adds[k]; !ok {
			deleteLen++
		}
	}
	return len(c.Adds) - deleteLen
}

// toMapState converts a single filter into a MapState entries added to 'p.PolicyMapState'.
//
// Note: It is possible for two selectors to select the same security ID.  To give priority to deny,
// AuthType, and L7 redirection (e.g., for visibility purposes), the mapstate entries are added to
// 'p.PolicyMapState' using insertWithChanges().
// Keys and old values of any added or deleted entries are added to 'changes'.
// 'redirects' is the map of currently realized redirects, it is used to find the proxy port for any redirects.
// p.SelectorCache is used as Identities interface during this call, which only has GetPrefix() that
// needs no lock.
func (l4 *L4Filter) toMapState(p *EndpointPolicy, features policyFeatures, changes ChangeState) {
	port := l4.Port
	proto := l4.U8Proto

	direction := trafficdirection.Egress
	if l4.Ingress {
		direction = trafficdirection.Ingress
	}

	logger := log
	if option.Config.Debug {
		logger = log.WithFields(logrus.Fields{
			logfields.EndpointID:       p.PolicyOwner.GetID(),
			logfields.Port:             port,
			logfields.PortName:         l4.PortName,
			logfields.Protocol:         proto,
			logfields.TrafficDirection: direction,
		})
	}

	// resolve named port
	if port == 0 && l4.PortName != "" {
		port = p.PolicyOwner.GetNamedPort(l4.Ingress, l4.PortName, proto)
		if port == 0 {
			return // nothing to be done for undefined named port
		}
	}

	var keysToAdd []Key
	for _, mp := range PortRangeToMaskedPorts(port, l4.EndPort) {
		keysToAdd = append(keysToAdd,
			KeyForDirection(direction).WithPortProtoPrefix(proto, mp.port, uint8(bits.LeadingZeros16(^mp.mask))))
	}

	// find the L7 rules for the wildcard entry, if any
	var wildcardRule *PerSelectorPolicy
	if l4.wildcard != nil {
		wildcardRule = l4.PerSelectorPolicies[l4.wildcard]
	}

	isL4Wildcard := (l4.Port != 0 || l4.PortName != "") && l4.wildcard != nil
	for cs, currentRule := range l4.PerSelectorPolicies {
		// have wildcard and this is an L3L4 key?
		isL3L4withWildcardPresent := isL4Wildcard && cs != l4.wildcard

		if isL3L4withWildcardPresent && wildcardRule.covers(currentRule) {
			logger.WithField(logfields.EndpointSelector, cs).Debug("ToMapState: Skipping L3/L4 key due to existing L4-only key")
			continue
		}

		isDenyRule := currentRule != nil && currentRule.IsDeny
		isRedirect := currentRule.IsRedirect()
		listener := currentRule.GetListener()
		priority := currentRule.GetPriority()
		if !isDenyRule && isL3L4withWildcardPresent && !isRedirect {
			// Inherit the redirect status from the wildcard rule.
			// This is now needed as 'covers()' can pass non-redirect L3L4 rules
			// that must inherit the redirect status from the L4-only (== L3-wildcard)
			// rule due to auth type on the L3L4 rule being different than in the
			// L4-only rule.
			isRedirect = wildcardRule.IsRedirect()
			listener = wildcardRule.GetListener()
			priority = wildcardRule.GetPriority()
		}
		authReq := currentRule.getAuthRequirement()
		var proxyPort uint16
		if isRedirect {
			var err error
			proxyPort, err = p.LookupRedirectPort(l4.Ingress, string(l4.Protocol), port, listener)
			if err != nil {
				// Skip unrealized redirects; this happens routineously just
				// before new redirects are realized. Once created, we are called
				// again.
				logger.WithError(err).WithField(logfields.EndpointSelector, cs).Debugf("Skipping unrealized redirect")
				continue
			}
		}
		entry := newMapStateEntry(l4.RuleOrigin[cs], proxyPort, priority, isDenyRule, authReq)

		if cs.IsWildcard() {
			for _, keyToAdd := range keysToAdd {
				keyToAdd.Identity = 0
				p.policyMapState.insertWithChanges(keyToAdd, entry, features, changes)

				if port == 0 {
					// Allow-all
					logger.WithField(logfields.EndpointSelector, cs).Debug("ToMapState: allow all")
				} else {
					// L4 allow
					logger.WithField(logfields.EndpointSelector, cs).Debug("ToMapState: L4 allow all")
				}
			}
			continue
		}

		idents := cs.GetSelections(p.VersionHandle)
		if option.Config.Debug {
			if isDenyRule {
				logger.WithFields(logrus.Fields{
					logfields.Version:          p.VersionHandle,
					logfields.EndpointSelector: cs,
					logfields.PolicyID:         idents,
				}).Debug("ToMapState: Denied remote IDs")
			} else {
				logger.WithFields(logrus.Fields{
					logfields.Version:          p.VersionHandle,
					logfields.EndpointSelector: cs,
					logfields.PolicyID:         idents,
				}).Debug("ToMapState: Allowed remote IDs")
			}
		}
		for _, id := range idents {
			for _, keyToAdd := range keysToAdd {
				keyToAdd.Identity = id
				p.policyMapState.insertWithChanges(keyToAdd, entry, features, changes)
				// If Cilium is in dual-stack mode then the "World" identity
				// needs to be split into two identities to represent World
				// IPv6 and IPv4 traffic distinctly from one another.
				if id == identity.ReservedIdentityWorld && option.Config.IsDualStack() {
					keyToAdd.Identity = identity.ReservedIdentityWorldIPv4
					p.policyMapState.insertWithChanges(keyToAdd, entry, features, changes)
					keyToAdd.Identity = identity.ReservedIdentityWorldIPv6
					p.policyMapState.insertWithChanges(keyToAdd, entry, features, changes)
				}
			}
		}
	}
	if option.Config.Debug {
		log.WithFields(logrus.Fields{
			logfields.PolicyKeysAdded:   changes.Adds,
			logfields.PolicyKeysDeleted: changes.Deletes,
			logfields.PolicyEntriesOld:  changes.old,
		}).Debug("ToMapChange changes")
	}
}

// IdentitySelectionUpdated implements CachedSelectionUser interface
// This call is made from a single goroutine in FIFO order to keep add
// and delete events ordered properly. No locks are held.
//
// The caller is responsible for making sure the same identity is not
// present in both 'added' and 'deleted'.
func (l4 *L4Filter) IdentitySelectionUpdated(cs types.CachedSelector, added, deleted []identity.NumericIdentity) {
	log.WithFields(logrus.Fields{
		logfields.EndpointSelector: cs,
		logfields.AddedPolicyID:    added,
		logfields.DeletedPolicyID:  deleted,
	}).Debug("identities selected by L4Filter updated")

	// Skip updates on wildcard selectors, as datapath and L7
	// proxies do not need enumeration of all ids for L3 wildcard.
	// This mirrors the per-selector logic in toMapState().
	if cs.IsWildcard() {
		return
	}

	// Push endpoint policy changes.
	//
	// `l4.policy` is nil when the filter is detached so
	// that we could not push updates on an unstable policy.
	l4Policy := l4.policy.Load()
	if l4Policy != nil {
		l4Policy.AccumulateMapChanges(l4, cs, added, deleted)
	}
}

func (l4 *L4Filter) IdentitySelectionCommit(txn *versioned.Tx) {
	log.WithField(logfields.NewVersion, txn).Debug("identity selection updates done")

	// Push endpoint policy incremental sync.
	//
	// `l4.policy` is nil when the filter is detached so
	// that we could not push updates on an unstable policy.
	l4Policy := l4.policy.Load()
	if l4Policy != nil {
		l4Policy.SyncMapChanges(l4, txn)
	}
}

func (l4 *L4Filter) IsPeerSelector() bool {
	return true
}

func (l4 *L4Filter) cacheIdentitySelector(sel api.EndpointSelector, lbls labels.LabelArray, selectorCache *SelectorCache) CachedSelector {
	cs, added := selectorCache.AddIdentitySelector(l4, lbls, sel)
	if added {
		l4.PerSelectorPolicies[cs] = nil // no per-selector policy (yet)
	}
	return cs
}

func (l4 *L4Filter) cacheIdentitySelectors(selectors api.EndpointSelectorSlice, lbls labels.LabelArray, selectorCache *SelectorCache) {
	for _, sel := range selectors {
		l4.cacheIdentitySelector(sel, lbls, selectorCache)
	}
}

func (l4 *L4Filter) cacheFQDNSelectors(selectors api.FQDNSelectorSlice, lbls labels.LabelArray, selectorCache *SelectorCache) {
	for _, fqdnSel := range selectors {
		l4.cacheFQDNSelector(fqdnSel, lbls, selectorCache)
	}
}

func (l4 *L4Filter) cacheFQDNSelector(sel api.FQDNSelector, lbls labels.LabelArray, selectorCache *SelectorCache) types.CachedSelector {
	cs, added := selectorCache.AddFQDNSelector(l4, lbls, sel)
	if added {
		l4.PerSelectorPolicies[cs] = nil // no per-selector policy (yet)
	}
	return cs
}

// add L7 rules for all endpoints in the L7DataMap
func (l7 L7DataMap) addPolicyForSelector(rules *api.L7Rules, terminatingTLS, originatingTLS *TLSContext, auth *api.Authentication, deny bool, sni []string, listener string, priority uint8) {
	isRedirect := !deny && (listener != "" || terminatingTLS != nil || originatingTLS != nil || len(sni) > 0 || !rules.IsEmpty())
	for epsel := range l7 {
		l7policy := &PerSelectorPolicy{
			TerminatingTLS: terminatingTLS,
			OriginatingTLS: originatingTLS,
			Authentication: auth,
			IsDeny:         deny,
			ServerNames:    NewStringSet(sni),
			isRedirect:     isRedirect,
			Listener:       listener,
			Priority:       priority,
		}
		if rules != nil {
			l7policy.L7Rules = *rules
		}
		l7[epsel] = l7policy
	}
}

type TLSDirection string

const (
	TerminatingTLS TLSDirection = "terminating"
	OriginatingTLS TLSDirection = "originating"
)

// getCerts reads certificates out of the PolicyContext, reading from k8s or local files depending on config
// and puts the values into the relevant keys in the TLSContext. Note that if the returned TLSContext.FromFile is
// `false`, then this will be read from Kubernetes.
func (l4 *L4Filter) getCerts(policyCtx PolicyContext, tls *api.TLSContext, direction TLSDirection) (*TLSContext, error) {
	if tls == nil {
		return nil, nil
	}
	ca, public, private, inlineSecrets, err := policyCtx.GetTLSContext(tls)
	if err != nil {
		log.WithError(err).Warningf("policy: Error getting %s TLS Context.", direction)
		return nil, err
	}

	// If the secret is not being included into NPDS inline, we're going to pass an SDS reference instead.
	if inlineSecrets {
		switch direction {
		case TerminatingTLS:
			if public == "" || private == "" {
				return nil, fmt.Errorf("Terminating TLS context is missing certs.")
			}
		case OriginatingTLS:
			if ca == "" {
				return nil, fmt.Errorf("Originating TLS context is missing CA certs.")
			}
		default:
			return nil, fmt.Errorf("invalid TLS direction: %s", direction)
		}
	} else {
		log.Debug("Secret being read from Kubernetes", "secret", k8sTypes.NamespacedName(*tls.Secret))
	}

	return &TLSContext{
		TrustedCA:        ca,
		CertificateChain: public,
		PrivateKey:       private,
		FromFile:         inlineSecrets,
		Secret:           k8sTypes.NamespacedName(*tls.Secret),
	}, nil
}

// createL4Filter creates a filter for L4 policy that applies to the specified
// endpoints and port/protocol, with reference to the original rules that the
// filter is derived from. This filter may be associated with a series of L7
// rules via the `rule` parameter.
// Not called with an empty peerEndpoints.
func createL4Filter(policyCtx PolicyContext, peerEndpoints api.EndpointSelectorSlice, auth *api.Authentication, rule api.Ports, port api.PortProtocol,
	protocol api.L4Proto, ruleLabels labels.LabelArray, ingress bool, fqdns api.FQDNSelectorSlice,
) (*L4Filter, error) {
	selectorCache := policyCtx.GetSelectorCache()

	portName := ""
	p := uint64(0)
	if iana.IsSvcName(port.Port) {
		portName = port.Port
	} else {
		// already validated via PortRule.Validate()
		p, _ = strconv.ParseUint(port.Port, 0, 16)
	}

	// already validated via L4Proto.Validate(), never "ANY"
	// NOTE: "ANY" for wildcarded port/proto!
	u8p, _ := u8proto.ParseProtocol(string(protocol))

	l4 := &L4Filter{
		Port:                uint16(p),            // 0 for L3-only rules and named ports
		EndPort:             uint16(port.EndPort), // 0 for a single port, >= 'Port' for a range
		PortName:            portName,             // non-"" for named ports
		Protocol:            protocol,
		U8Proto:             u8p,
		PerSelectorPolicies: make(L7DataMap),
		RuleOrigin:          make(map[CachedSelector]labels.LabelArrayList), // Filled in below.
		Ingress:             ingress,
	}

	if peerEndpoints.SelectsAllEndpoints() {
		l4.wildcard = l4.cacheIdentitySelector(api.WildcardEndpointSelector, ruleLabels, selectorCache)
	} else {
		l4.cacheIdentitySelectors(peerEndpoints, ruleLabels, selectorCache)
		l4.cacheFQDNSelectors(fqdns, ruleLabels, selectorCache)
	}

	var terminatingTLS *TLSContext
	var originatingTLS *TLSContext
	var rules *api.L7Rules
	var sni []string
	listener := ""
	var priority uint8

	pr := rule.GetPortRule()
	if pr != nil {
		rules = pr.Rules
		sni = pr.ServerNames

		// Get TLS contexts, if any
		var err error
		terminatingTLS, err = l4.getCerts(policyCtx, pr.TerminatingTLS, TerminatingTLS)
		if err != nil {
			return nil, err
		}
		originatingTLS, err = l4.getCerts(policyCtx, pr.OriginatingTLS, OriginatingTLS)
		if err != nil {
			return nil, err
		}

		// Set parser type to TLS, if TLS. This will be overridden by L7 below, if rules
		// exists.
		if terminatingTLS != nil || originatingTLS != nil || len(pr.ServerNames) > 0 {
			l4.L7Parser = ParserTypeTLS
		}

		// Determine L7ParserType from rules present. Earlier validation ensures rules
		// for multiple protocols are not present here.
		if rules != nil {
			// we need this to redirect DNS UDP (or ANY, which is more useful)
			if len(rules.DNS) > 0 {
				l4.L7Parser = ParserTypeDNS
			} else if protocol == api.ProtoTCP { // Other than DNS only support TCP
				switch {
				case len(rules.HTTP) > 0:
					l4.L7Parser = ParserTypeHTTP
				case len(rules.Kafka) > 0:
					l4.L7Parser = ParserTypeKafka
				case rules.L7Proto != "":
					l4.L7Parser = (L7ParserType)(rules.L7Proto)
				}
			}
		}

		// Override the parser type to CRD is applicable.
		if pr.Listener != nil {
			l4.L7Parser = ParserTypeCRD
			ns := policyCtx.GetNamespace()
			resource := pr.Listener.EnvoyConfig
			switch resource.Kind {
			case "CiliumEnvoyConfig":
				if ns == "" {
					// Cluster-scoped CCNP tries to use namespaced
					// CiliumEnvoyConfig
					//
					// TODO: Catch this in rule validation once we have a
					// validation context in there so that we can differentiate
					// between CNP and CCNP at validation time.
					return nil, fmt.Errorf("Listener %q in CCNP can not use Kind CiliumEnvoyConfig", pr.Listener.Name)
				}
			case "CiliumClusterwideEnvoyConfig":
				// CNP refers to a cluster-scoped listener
				ns = ""
			default:
			}
			listener, _ = api.ResourceQualifiedName(ns, resource.Name, pr.Listener.Name, api.ForceNamespace)
			priority = pr.Listener.Priority
		}
	}

	if l4.L7Parser != ParserTypeNone || auth != nil || policyCtx.IsDeny() {
		modifiedRules := rules

		// If we have L7 rules and default deny is disabled (EnableDefaultDeny=false), we should ensure those rules
		// don't cause other L7 traffic to be denied.
		// Special handling for L7 rules is applied when:
		// 1. We have L7 rules
		// 2. Default deny is disabled for this direction
		// 3. This is a positive policy (not a deny policy)
		hasL7Rules := !rules.IsEmpty()
		isDefaultDenyDisabled := (ingress && !policyCtx.DefaultDenyIngress()) || (!ingress && !policyCtx.DefaultDenyEgress())
		isAllowPolicy := !policyCtx.IsDeny()

		if hasL7Rules && isDefaultDenyDisabled && isAllowPolicy {
			log.WithFields(logrus.Fields{
				logfields.L7Parser: l4.L7Parser,
				logfields.Ingress:  ingress,
			}).Debug("Adding wildcard L7 rules for default-allow policy")

			modifiedRules = ensureWildcard(rules, l4.L7Parser)
		}

		l4.PerSelectorPolicies.addPolicyForSelector(modifiedRules, terminatingTLS, originatingTLS, auth, policyCtx.IsDeny(), sni, listener, priority)
	}

	for cs := range l4.PerSelectorPolicies {
		l4.RuleOrigin[cs] = labels.LabelArrayList{ruleLabels}
	}

	return l4, nil
}

func (l4 *L4Filter) removeSelectors(selectorCache *SelectorCache) {
	selectors := make(types.CachedSelectorSlice, 0, len(l4.PerSelectorPolicies))
	for cs := range l4.PerSelectorPolicies {
		selectors = append(selectors, cs)
	}
	selectorCache.RemoveSelectors(selectors, l4)
}

// detach releases the references held in the L4Filter and must be called before
// the filter is left to be garbage collected.
// L4Filter may still be accessed concurrently after it has been detached.
func (l4 *L4Filter) detach(selectorCache *SelectorCache) {
	l4.removeSelectors(selectorCache)
	l4.policy.Store(nil)
}

// attach signifies that the L4Filter is ready and reacheable for updates
// from SelectorCache. L4Filter (and L4Policy) is read-only after this is called,
// multiple goroutines will be reading the fields from that point on.
func (l4 *L4Filter) attach(ctx PolicyContext, l4Policy *L4Policy) policyFeatures {
	// All rules have been added to the L4Filter at this point.
	// Sort the rules label array list for more efficient equality comparison.
	for _, labels := range l4.RuleOrigin {
		labels.Sort()
	}

	var features policyFeatures
	for cs, cp := range l4.PerSelectorPolicies {
		if cp != nil {
			if cp.isRedirect {
				features.setFeature(redirectRules)
			}

			if cp.IsDeny {
				features.setFeature(denyRules)
			}

			explicit, authType := getAuthType(cp.Authentication)
			if explicit {
				features.setFeature(authRules)

				if authType != types.AuthTypeDisabled {
					if l4Policy.authMap == nil {
						l4Policy.authMap = make(authMap, 1)
					}
					authTypes := l4Policy.authMap[cs]
					if authTypes == nil {
						authTypes = make(AuthTypes, 1)
					}
					authTypes[authType] = struct{}{}
					l4Policy.authMap[cs] = authTypes
				}
			}

			// Compute Envoy policies when a policy is ready to be used
			if len(cp.L7Rules.HTTP) > 0 {
				cp.EnvoyHTTPRules, cp.CanShortCircuit = ctx.GetEnvoyHTTPRules(&cp.L7Rules)
			}
		}
	}

	l4.policy.Store(l4Policy)
	return features
}

// createL4IngressFilter creates a filter for L4 policy that applies to the
// specified endpoints and port/protocol for ingress traffic, with reference
// to the original rules that the filter is derived from. This filter may be
// associated with a series of L7 rules via the `rule` parameter.
//
// hostWildcardL7 determines if L7 traffic from Host should be
// wildcarded (in the relevant daemon mode).
func createL4IngressFilter(policyCtx PolicyContext, fromEndpoints api.EndpointSelectorSlice, auth *api.Authentication, hostWildcardL7 []string, rule api.Ports, port api.PortProtocol,
	protocol api.L4Proto, ruleLabels labels.LabelArray,
) (*L4Filter, error) {
	filter, err := createL4Filter(policyCtx, fromEndpoints, auth, rule, port, protocol, ruleLabels, true, nil)
	if err != nil {
		return nil, err
	}

	// If the filter would apply proxy redirection for the Host, when we should accept
	// everything from host, then wildcard Host at L7.
	if len(hostWildcardL7) > 0 {
		for cs, l7 := range filter.PerSelectorPolicies {
			if l7.IsRedirect() && cs.Selects(versioned.Latest(), identity.ReservedIdentityHost) {
				for _, name := range hostWildcardL7 {
					selector := api.ReservedEndpointSelectors[name]
					filter.cacheIdentitySelector(selector, ruleLabels, policyCtx.GetSelectorCache())
				}
			}
		}
	}

	return filter, nil
}

// createL4EgressFilter creates a filter for L4 policy that applies to the
// specified endpoints and port/protocol for egress traffic, with reference
// to the original rules that the filter is derived from. This filter may be
// associated with a series of L7 rules via the `rule` parameter.
func createL4EgressFilter(policyCtx PolicyContext, toEndpoints api.EndpointSelectorSlice, auth *api.Authentication, rule api.Ports, port api.PortProtocol,
	protocol api.L4Proto, ruleLabels labels.LabelArray, fqdns api.FQDNSelectorSlice,
) (*L4Filter, error) {
	return createL4Filter(policyCtx, toEndpoints, auth, rule, port, protocol, ruleLabels, false, fqdns)
}

// redirectType returns the redirectType for this filter
func (l4 *L4Filter) redirectType() redirectTypes {
	switch l4.L7Parser {
	case ParserTypeNone:
		return redirectTypeNone
	case ParserTypeDNS:
		return redirectTypeDNS
	case ParserTypeHTTP, ParserTypeTLS, ParserTypeCRD:
		return redirectTypeEnvoy
	default:
		// all other (non-empty) values are used for proxylib redirects
		return redirectTypeProxylib
	}
}

// IsRedirect returns true if the L4 filter contains a port redirection
func (l4 *L4Filter) IsRedirect() bool {
	return l4.L7Parser != ParserTypeNone
}

// Marshal returns the `L4Filter` in a JSON string.
func (l4 *L4Filter) Marshal() string {
	b, err := json.Marshal(l4)
	if err != nil {
		b = []byte("\"L4Filter error: " + err.Error() + "\"")
	}
	return string(b)
}

// String returns the `L4Filter` in a human-readable string.
func (l4 *L4Filter) String() string {
	b, err := json.Marshal(l4)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// Note: Only used for policy tracing
func (l4 *L4Filter) matchesLabels(labels labels.LabelArray) (bool, bool) {
	if l4.wildcard != nil {
		perSelectorPolicy := l4.PerSelectorPolicies[l4.wildcard]
		isDeny := perSelectorPolicy != nil && perSelectorPolicy.IsDeny
		return true, isDeny
	} else if len(labels) == 0 {
		return false, false
	}

	var selected bool
	for sel, rule := range l4.PerSelectorPolicies {
		// slow, but OK for tracing
		idSel := sel.(*identitySelector)
		if lis, ok := idSel.source.(*labelIdentitySelector); ok && lis.xxxMatches(labels) {
			isDeny := rule != nil && rule.IsDeny
			selected = true
			if isDeny {
				return true, isDeny
			}
		}
	}
	return selected, false
}

// addL4Filter adds 'filterToMerge' into the 'resMap'. Returns an error if it
// the 'filterToMerge' can't be merged with an existing filter for the same
// port and proto.
func addL4Filter(policyCtx PolicyContext,
	ctx *SearchContext, resMap L4PolicyMap,
	p api.PortProtocol, proto api.L4Proto,
	filterToMerge *L4Filter,
) error {
	existingFilter := resMap.ExactLookup(p.Port, uint16(p.EndPort), string(proto))
	if existingFilter == nil {
		resMap.Upsert(p.Port, uint16(p.EndPort), string(proto), filterToMerge)
		return nil
	}

	selectorCache := policyCtx.GetSelectorCache()
	if err := mergePortProto(ctx, existingFilter, filterToMerge, selectorCache); err != nil {
		filterToMerge.detach(selectorCache)
		return err
	}

	// To keep the rule origin tracking correct, merge the rule label arrays for each CachedSelector
	// we know about. New CachedSelectors are added.
	for cs, newLabels := range filterToMerge.RuleOrigin {
		if existingLabels, ok := existingFilter.RuleOrigin[cs]; ok {
			existingFilter.RuleOrigin[cs] = existingLabels.MergeSorted(newLabels)
		} else {
			existingFilter.RuleOrigin[cs] = newLabels
		}
	}

	resMap.Upsert(p.Port, uint16(p.EndPort), string(proto), existingFilter)
	return nil
}

// L4PolicyMap is a list of L4 filters indexable by port/endport/protocol
type L4PolicyMap interface {
	Upsert(port string, endPort uint16, protocol string, l4 *L4Filter)
	Delete(port string, endPort uint16, protocol string)
	ExactLookup(port string, endPort uint16, protocol string) *L4Filter
	MatchesLabels(port, protocol string, labels labels.LabelArray) (match, isDeny bool)
	Detach(selectorCache *SelectorCache)
	IngressCoversContext(ctx *SearchContext) api.Decision
	EgressCoversContext(ctx *SearchContext) api.Decision
	ForEach(func(l4 *L4Filter) bool)
	TestingOnlyEquals(bMap L4PolicyMap) bool
	TestingOnlyDiff(expectedMap L4PolicyMap) string
	Len() int
}

// NewL4PolicyMap creates an new L4PolicMap.
func NewL4PolicyMap() L4PolicyMap {
	return &l4PolicyMap{
		namedPortMap:   make(map[string]*L4Filter),
		rangePortMap:   make(map[portProtoKey]*L4Filter),
		rangePortIndex: bitlpm.NewUintTrie[uint32, map[portProtoKey]struct{}](),
	}
}

// NewL4PolicyMapWithValues creates an new L4PolicMap, with an initial
// set of values. The initMap argument does not support port ranges.
func NewL4PolicyMapWithValues(initMap map[string]*L4Filter) L4PolicyMap {
	l4M := &l4PolicyMap{
		namedPortMap:   make(map[string]*L4Filter),
		rangePortMap:   make(map[portProtoKey]*L4Filter),
		rangePortIndex: bitlpm.NewUintTrie[uint32, map[portProtoKey]struct{}](),
	}
	for k, v := range initMap {
		portProtoSlice := strings.Split(k, "/")
		if len(portProtoSlice) < 2 {
			continue
		}
		l4M.Upsert(portProtoSlice[0], 0, portProtoSlice[1], v)
	}
	return l4M
}

type portProtoKey struct {
	port, endPort uint16
	proto         uint8
}

// l4PolicyMap is the implementation of L4PolicyMap
type l4PolicyMap struct {
	// namedPortMap represents the named ports (a Kubernetes feature)
	// that map to an L4Filter. They must be tracked at the selection
	// level, because they can only be resolved at the endpoint/identity
	// level. Named ports cannot have ranges.
	namedPortMap map[string]*L4Filter
	// rangePortMap is a map of all L4Filters indexed by their port-
	// protocol.
	rangePortMap map[portProtoKey]*L4Filter
	// rangePortIndex is an index of all L4Filters so that
	// L4Filters that have overlapping port ranges can be looked up
	// by with a single port.
	rangePortIndex *bitlpm.UintTrie[uint32, map[portProtoKey]struct{}]
}

func parsePortProtocol(port, protocol string) (uint16, uint8) {
	// These string values have been validated many times
	// over at this point.
	prt, _ := strconv.ParseUint(port, 10, 16)
	proto, _ := u8proto.ParseProtocol(protocol)
	return uint16(prt), uint8(proto)
}

// makePolicyMapKey creates a protocol-port uint32 with the
// upper 16 bits containing the protocol and the lower 16
// bits containing the port.
func makePolicyMapKey(port, mask uint16, proto uint8) uint32 {
	return (uint32(proto) << 16) | uint32(port&mask)
}

// Upsert L4Filter adds an L4Filter indexed by protocol/port-endPort.
func (l4M *l4PolicyMap) Upsert(port string, endPort uint16, protocol string, l4 *L4Filter) {
	if iana.IsSvcName(port) {
		l4M.namedPortMap[port+"/"+protocol] = l4
		return
	}

	portU, protoU := parsePortProtocol(port, protocol)
	ppK := portProtoKey{
		port:    portU,
		endPort: endPort,
		proto:   protoU,
	}
	_, indexExists := l4M.rangePortMap[ppK]
	l4M.rangePortMap[ppK] = l4
	// We do not need to reindex a key that already exists,
	// even if the filter changed.
	if !indexExists {
		for _, mp := range PortRangeToMaskedPorts(portU, endPort) {
			k := makePolicyMapKey(mp.port, mp.mask, protoU)
			prefix := 32 - uint(bits.TrailingZeros16(mp.mask))
			portProtoSet, ok := l4M.rangePortIndex.ExactLookup(prefix, k)
			if !ok {
				portProtoSet = make(map[portProtoKey]struct{})
				l4M.rangePortIndex.Upsert(prefix, k, portProtoSet)
			}
			portProtoSet[ppK] = struct{}{}
		}
	}
}

// Delete an L4Filter from the index by protocol/port-endPort
func (l4M *l4PolicyMap) Delete(port string, endPort uint16, protocol string) {
	if iana.IsSvcName(port) {
		delete(l4M.namedPortMap, port+"/"+protocol)
		return
	}

	portU, protoU := parsePortProtocol(port, protocol)
	ppK := portProtoKey{
		port:    portU,
		endPort: endPort,
		proto:   protoU,
	}
	_, indexExists := l4M.rangePortMap[ppK]
	delete(l4M.rangePortMap, ppK)
	// Only delete the index if the key exists.
	if indexExists {
		for _, mp := range PortRangeToMaskedPorts(portU, endPort) {
			k := makePolicyMapKey(mp.port, mp.mask, protoU)
			prefix := 32 - uint(bits.TrailingZeros16(mp.mask))
			portProtoSet, ok := l4M.rangePortIndex.ExactLookup(prefix, k)
			if !ok {
				return
			}
			delete(portProtoSet, ppK)
			if len(portProtoSet) == 0 {
				l4M.rangePortIndex.Delete(prefix, k)
			}
		}
	}
}

// ExactLookup looks up an L4Filter by protocol/port-endPort and looks for an exact match.
func (l4M *l4PolicyMap) ExactLookup(port string, endPort uint16, protocol string) *L4Filter {
	if iana.IsSvcName(port) {
		return l4M.namedPortMap[port+"/"+protocol]
	}

	portU, protoU := parsePortProtocol(port, protocol)
	ppK := portProtoKey{
		port:    portU,
		endPort: endPort,
		proto:   protoU,
	}
	return l4M.rangePortMap[ppK]
}

// MatchesLabels checks if a given port, protocol, and labels matches
// any Rule in the L4PolicyMap.
func (l4M *l4PolicyMap) MatchesLabels(port, protocol string, labels labels.LabelArray) (match, isDeny bool) {
	if iana.IsSvcName(port) {
		l4 := l4M.namedPortMap[port+"/"+protocol]
		if l4 != nil {
			return l4.matchesLabels(labels)
		}
		return
	}

	portU, protoU := parsePortProtocol(port, protocol)
	l4PortProtoKeys := make(map[portProtoKey]struct{})
	l4M.rangePortIndex.Ancestors(32, makePolicyMapKey(portU, 0xffff, protoU),
		func(_ uint, _ uint32, portProtoSet map[portProtoKey]struct{}) bool {
			for k := range portProtoSet {
				v, ok := l4M.rangePortMap[k]
				if ok {
					if _, ok := l4PortProtoKeys[k]; !ok {
						match, isDeny = v.matchesLabels(labels)
						if isDeny {
							return false
						}
					}
				}
			}
			return true
		})
	return
}

// ForEach iterates over all L4Filters in the l4PolicyMap.
func (l4M *l4PolicyMap) ForEach(fn func(l4 *L4Filter) bool) {
	for _, f := range l4M.namedPortMap {
		if !fn(f) {
			return
		}
	}
	for _, v := range l4M.rangePortMap {
		if !fn(v) {
			return
		}
	}
}

// Equals returns true if both L4PolicyMaps are equal.
func (l4M *l4PolicyMap) TestingOnlyEquals(bMap L4PolicyMap) bool {
	if l4M.Len() != bMap.Len() {
		return false
	}
	equal := true
	l4M.ForEach(func(l4 *L4Filter) bool {
		port := l4.PortName
		if len(port) == 0 {
			port = fmt.Sprintf("%d", l4.Port)
		}
		l4B := bMap.ExactLookup(port, l4.EndPort, string(l4.Protocol))
		equal = l4.Equals(l4B)
		return equal
	})
	return equal
}

// Diff returns the difference between to L4PolicyMaps.
func (l4M *l4PolicyMap) TestingOnlyDiff(expected L4PolicyMap) (res string) {
	res += "Missing (-), Unexpected (+):\n"
	expected.ForEach(func(eV *L4Filter) bool {
		port := eV.PortName
		if len(port) == 0 {
			port = fmt.Sprintf("%d", eV.Port)
		}
		oV := l4M.ExactLookup(port, eV.Port, string(eV.Protocol))
		if oV != nil {
			if !eV.Equals(oV) {
				res += "- " + eV.String() + "\n"
				res += "+ " + oV.String() + "\n"
			}
		} else {
			res += "- " + eV.String() + "\n"
		}
		return true
	})
	l4M.ForEach(func(oV *L4Filter) bool {
		port := oV.PortName
		if len(port) == 0 {
			port = fmt.Sprintf("%d", oV.Port)
		}
		eV := expected.ExactLookup(port, oV.Port, string(oV.Protocol))
		if eV == nil {
			res += "+ " + oV.String() + "\n"
		}
		return true
	})
	return
}

// Len returns the number of entries in the map.
func (l4M *l4PolicyMap) Len() int {
	if l4M == nil {
		return 0
	}
	return len(l4M.namedPortMap) + len(l4M.rangePortMap)
}

type policyFeatures uint8

const (
	denyRules policyFeatures = 1 << iota
	redirectRules
	authRules

	allFeatures policyFeatures = ^policyFeatures(0)
)

func (pf *policyFeatures) setFeature(feature policyFeatures) {
	*pf |= feature
}

func (pf policyFeatures) contains(feature policyFeatures) bool {
	return pf&feature != 0
}

type L4DirectionPolicy struct {
	PortRules L4PolicyMap

	// features tracks properties of PortRules to skip code when features are not used
	features policyFeatures
}

func newL4DirectionPolicy() L4DirectionPolicy {
	return L4DirectionPolicy{
		PortRules: NewL4PolicyMap(),
	}
}

// Detach removes the cached selectors held by L4PolicyMap from the
// selectorCache, allowing the map to be garbage collected when there
// are no more references to it.
func (l4 L4DirectionPolicy) Detach(selectorCache *SelectorCache) {
	l4.PortRules.Detach(selectorCache)
}

// detach is used directly from tracing and testing functions
func (l4M *l4PolicyMap) Detach(selectorCache *SelectorCache) {
	l4M.ForEach(func(l4 *L4Filter) bool {
		l4.detach(selectorCache)
		return true
	})
}

// Attach makes all the L4Filters to point back to the L4Policy that contains them.
// This is done before the L4PolicyMap is exposed to concurrent access.
// Returns the bitmask of all redirect types for this policymap.
func (l4 *L4DirectionPolicy) attach(ctx PolicyContext, l4Policy *L4Policy) redirectTypes {
	var redirectTypes redirectTypes
	var features policyFeatures
	l4.PortRules.ForEach(func(f *L4Filter) bool {
		features |= f.attach(ctx, l4Policy)
		redirectTypes |= f.redirectType()
		return true
	})
	l4.features = features
	return redirectTypes
}

// containsAllL3L4 checks if the L4PolicyMap contains all L4 ports in `ports`.
// For L4Filters that specify ToEndpoints or FromEndpoints, uses `labels` to
// determine whether the policy allows L4 communication between the corresponding
// endpoints.
// Returns api.Denied in the following conditions:
//   - If a single port is not present in the `L4PolicyMap` and is not allowed
//     by the distilled L3 policy
//   - If a port is present in the `L4PolicyMap`, but it applies ToEndpoints or
//     FromEndpoints constraints that require labels not present in `labels`.
//
// Otherwise, returns api.Allowed.
//
// Note: Only used for policy tracing
func (l4M *l4PolicyMap) containsAllL3L4(labels labels.LabelArray, ports []*models.Port) api.Decision {
	if l4M.Len() == 0 {
		return api.Allowed
	}

	// Check L3-only filters first.
	filter := l4M.ExactLookup("0", 0, "ANY")
	if filter != nil {

		matches, isDeny := filter.matchesLabels(labels)
		switch {
		case matches && isDeny:
			return api.Denied
		case matches:
			return api.Allowed
		}
	}

	for _, l4Ctx := range ports {
		portStr := l4Ctx.Name
		if !iana.IsSvcName(portStr) {
			portStr = strconv.FormatUint(uint64(l4Ctx.Port), 10)
		}
		lwrProtocol := l4Ctx.Protocol
		switch lwrProtocol {
		case "", models.PortProtocolANY:
			tcpmatch, isTCPDeny := l4M.MatchesLabels(portStr, "TCP", labels)
			udpmatch, isUDPDeny := l4M.MatchesLabels(portStr, "UDP", labels)
			sctpmatch, isSCTPDeny := l4M.MatchesLabels(portStr, "SCTP", labels)
			if (!tcpmatch && !udpmatch && !sctpmatch) || (isTCPDeny && isUDPDeny && isSCTPDeny) {
				return api.Denied
			}
		default:
			matches, isDeny := l4M.MatchesLabels(portStr, lwrProtocol, labels)
			if !matches || isDeny {
				return api.Denied
			}
		}
	}
	return api.Allowed
}

type L4Policy struct {
	Ingress L4DirectionPolicy
	Egress  L4DirectionPolicy

	authMap authMap

	// Revision is the repository revision used to generate this policy.
	Revision uint64

	// redirectTypes is a bitmap containing the types of redirect contained by this policy.  It
	// is computed after the policy maps to avoid scanning them repeatedly when using the
	// L4Policy
	redirectTypes redirectTypes

	// Endpoint policies using this L4Policy
	// These are circular references, cleaned up in Detach()
	// This mutex is taken while Endpoint mutex is held, so Endpoint lock
	// MUST always be taken before this mutex.
	mutex lock.RWMutex
	users map[*EndpointPolicy]struct{}
}

// NewL4Policy creates a new L4Policy
func NewL4Policy(revision uint64) L4Policy {
	return L4Policy{
		Ingress:  newL4DirectionPolicy(),
		Egress:   newL4DirectionPolicy(),
		Revision: revision,
		users:    make(map[*EndpointPolicy]struct{}),
	}
}

// insertUser adds a user to the L4Policy so that incremental
// updates of the L4Policy may be forwarded to the users of it.
// May not call into SelectorCache, as SelectorCache is locked during this call.
func (l4 *L4Policy) insertUser(user *EndpointPolicy) {
	l4.mutex.Lock()

	// 'users' is set to nil when the policy is detached. This
	// happens to the old policy when it is being replaced with a
	// new one, or when the last endpoint using this policy is
	// removed.
	// In the case of an policy update it is possible that an
	// endpoint has started regeneration before the policy was
	// updated, and that the policy was updated before the said
	// endpoint reached this point. In this case the endpoint's
	// policy is going to be recomputed soon after and we do
	// nothing here.
	if l4.users != nil {
		l4.users[user] = struct{}{}
	}

	l4.mutex.Unlock()
}

// removeUser removes a user that no longer needs incremental updates
// from the L4Policy.
func (l4 *L4Policy) removeUser(user *EndpointPolicy) {
	// 'users' is set to nil when the policy is detached. This
	// happens to the old policy when it is being replaced with a
	// new one, or when the last endpoint using this policy is
	// removed.
	l4.mutex.Lock()
	if l4.users != nil {
		delete(l4.users, user)
	}
	l4.mutex.Unlock()
}

// AccumulateMapChanges distributes the given changes to the registered users.
//
// The caller is responsible for making sure the same identity is not
// present in both 'adds' and 'deletes'.
func (l4Policy *L4Policy) AccumulateMapChanges(l4 *L4Filter, cs CachedSelector, adds, deletes []identity.NumericIdentity) {
	port := uint16(l4.Port)
	proto := l4.U8Proto
	derivedFrom := l4.RuleOrigin[cs]

	direction := trafficdirection.Egress
	if l4.Ingress {
		direction = trafficdirection.Ingress
	}
	perSelectorPolicy := l4.PerSelectorPolicies[cs]
	redirect := perSelectorPolicy.IsRedirect()
	listener := perSelectorPolicy.GetListener()
	priority := perSelectorPolicy.GetPriority()
	authReq := perSelectorPolicy.getAuthRequirement()
	isDeny := perSelectorPolicy != nil && perSelectorPolicy.IsDeny

	// Can hold rlock here as neither GetNamedPort() nor LookupRedirectPort() no longer
	// takes the Endpoint lock below.
	// SelectorCache may not be called into while holding this lock!
	l4Policy.mutex.RLock()
	defer l4Policy.mutex.RUnlock()

	for epPolicy := range l4Policy.users {
		// resolve named port
		if port == 0 && l4.PortName != "" {
			port = epPolicy.PolicyOwner.GetNamedPort(l4.Ingress, l4.PortName, proto)
			if port == 0 {
				continue
			}
		}
		var proxyPort uint16
		if redirect {
			var err error
			proxyPort, err = epPolicy.LookupRedirectPort(l4.Ingress, string(l4.Protocol), port, listener)
			if err != nil {
				log.WithFields(logrus.Fields{
					logfields.EndpointSelector: cs,
					logfields.Port:             port,
					logfields.Protocol:         proto,
					logfields.TrafficDirection: direction,
					logfields.IsRedirect:       redirect,
					logfields.Listener:         listener,
					logfields.ListenerPriority: priority,
				}).Warn("AccumulateMapChanges: Missing redirect.")
				continue
			}
		}
		var keysToAdd []Key
		for _, mp := range PortRangeToMaskedPorts(port, l4.EndPort) {
			keysToAdd = append(keysToAdd,
				KeyForDirection(direction).WithPortProtoPrefix(proto, mp.port, uint8(bits.LeadingZeros16(^mp.mask))))
		}
		value := newMapStateEntry(derivedFrom, proxyPort, priority, isDeny, authReq)

		if option.Config.Debug {
			authString := "default"
			if authReq.IsExplicit() {
				authString = authReq.AuthType().String()
			}
			log.WithFields(logrus.Fields{
				logfields.EndpointSelector: cs,
				logfields.AddedPolicyID:    adds,
				logfields.DeletedPolicyID:  deletes,
				logfields.Port:             port,
				logfields.Protocol:         proto,
				logfields.TrafficDirection: direction,
				logfields.IsRedirect:       redirect,
				logfields.AuthType:         authString,
				logfields.Listener:         listener,
				logfields.ListenerPriority: priority,
			}).Debug("AccumulateMapChanges")
		}
		epPolicy.policyMapChanges.AccumulateMapChanges(adds, deletes, keysToAdd, value)
	}
}

// SyncMapChanges marks earlier updates as completed
func (l4Policy *L4Policy) SyncMapChanges(l4 *L4Filter, txn *versioned.Tx) {
	// SelectorCache may not be called into while holding this lock!
	l4Policy.mutex.RLock()

	for epPolicy := range l4Policy.users {
		epPolicy.policyMapChanges.SyncMapChanges(txn)
	}
	l4Policy.mutex.RUnlock()
}

// Detach makes the L4Policy ready for garbage collection, removing
// circular pointer references.
// Note that the L4Policy itself is not modified in any way, so that it may still
// be used concurrently.
func (l4 *L4Policy) Detach(selectorCache *SelectorCache) {
	l4.Ingress.Detach(selectorCache)
	l4.Egress.Detach(selectorCache)

	l4.mutex.Lock()
	l4.users = nil
	l4.mutex.Unlock()
}

// Attach makes all the L4Filters to point back to the L4Policy that contains them.
// This is done before the L4Policy is exposed to concurrent access.
func (l4 *L4Policy) Attach(ctx PolicyContext) {
	ingressRedirects := l4.Ingress.attach(ctx, l4)
	egressRedirects := l4.Egress.attach(ctx, l4)
	l4.redirectTypes = ingressRedirects | egressRedirects
}

// IngressCoversContext checks if the receiver's ingress L4Policy contains
// all `dPorts` and `labels`.
//
// Note: Only used for policy tracing
func (l4M *l4PolicyMap) IngressCoversContext(ctx *SearchContext) api.Decision {
	return l4M.containsAllL3L4(ctx.From, ctx.DPorts)
}

// EgressCoversContext checks if the receiver's egress L4Policy contains
// all `dPorts` and `labels`.
//
// Note: Only used for policy tracing
func (l4M *l4PolicyMap) EgressCoversContext(ctx *SearchContext) api.Decision {
	return l4M.containsAllL3L4(ctx.To, ctx.DPorts)
}

// HasRedirect returns true if the L4 policy contains at least one port redirection
func (l4 *L4Policy) HasRedirect() bool {
	return l4 != nil && l4.redirectTypes != redirectTypeNone
}

// HasEnvoyRedirect returns true if the L4 policy contains at least one port redirection to Envoy
func (l4 *L4Policy) HasEnvoyRedirect() bool {
	return l4 != nil && l4.redirectTypes&redirectTypeEnvoy == redirectTypeEnvoy
}

// HasProxylibRedirect returns true if the L4 policy contains at least one port redirection to Proxylib
func (l4 *L4Policy) HasProxylibRedirect() bool {
	return l4 != nil && l4.redirectTypes&redirectTypeProxylib == redirectTypeProxylib
}

func (l4 *L4Policy) GetModel() *models.L4Policy {
	if l4 == nil {
		return nil
	}

	ingress := []*models.PolicyRule{}
	l4.Ingress.PortRules.ForEach(func(v *L4Filter) bool {
		rulesBySelector := map[string][][]string{}
		derivedFrom := labels.LabelArrayList{}
		for sel, rules := range v.RuleOrigin {
			derivedFrom.MergeSorted(rules)
			rulesBySelector[sel.String()] = rules.GetModel()
		}
		ingress = append(ingress, &models.PolicyRule{
			Rule:             v.Marshal(),
			DerivedFromRules: derivedFrom.GetModel(),
			RulesBySelector:  rulesBySelector,
		})
		return true
	})

	egress := []*models.PolicyRule{}
	l4.Egress.PortRules.ForEach(func(v *L4Filter) bool {
		derivedFrom := labels.LabelArrayList{}
		for _, rules := range v.RuleOrigin {
			derivedFrom.MergeSorted(rules)
		}
		egress = append(egress, &models.PolicyRule{
			Rule:             v.Marshal(),
			DerivedFromRules: derivedFrom.GetModel(),
		})
		return true
	})

	return &models.L4Policy{
		Ingress: ingress,
		Egress:  egress,
	}
}

// ProxyPolicy is any type which encodes state needed to redirect to an L7
// proxy.
type ProxyPolicy interface {
	CopyL7RulesPerEndpoint() L7DataMap
	GetL7Parser() L7ParserType
	GetIngress() bool
	GetPort() uint16
	GetProtocol() u8proto.U8proto
	GetListener() string
}
