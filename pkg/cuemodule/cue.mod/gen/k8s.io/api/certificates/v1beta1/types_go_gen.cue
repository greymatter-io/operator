// Code generated by cue get go. DO NOT EDIT.

//cue:generate cue get go k8s.io/api/certificates/v1beta1

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"
)

// Describes a certificate signing request
#CertificateSigningRequest: {
	metav1.#TypeMeta

	// +optional
	metadata?: metav1.#ObjectMeta @go(ObjectMeta) @protobuf(1,bytes,opt)

	// spec contains the certificate request, and is immutable after creation.
	// Only the request, signerName, expirationSeconds, and usages fields can be set on creation.
	// Other fields are derived by Kubernetes and cannot be modified by users.
	spec: #CertificateSigningRequestSpec @go(Spec) @protobuf(2,bytes,opt)

	// Derived information about the request.
	// +optional
	status?: #CertificateSigningRequestStatus @go(Status) @protobuf(3,bytes,opt)
}

// CertificateSigningRequestSpec contains the certificate request.
#CertificateSigningRequestSpec: {
	// Base64-encoded PKCS#10 CSR data
	// +listType=atomic
	request: bytes @go(Request,[]byte) @protobuf(1,bytes,opt)

	// Requested signer for the request. It is a qualified name in the form:
	// `scope-hostname.io/name`.
	// If empty, it will be defaulted:
	//  1. If it's a kubelet client certificate, it is assigned
	//     "kubernetes.io/kube-apiserver-client-kubelet".
	//  2. If it's a kubelet serving certificate, it is assigned
	//     "kubernetes.io/kubelet-serving".
	//  3. Otherwise, it is assigned "kubernetes.io/legacy-unknown".
	// Distribution of trust for signers happens out of band.
	// You can select on this field using `spec.signerName`.
	// +optional
	signerName?: null | string @go(SignerName,*string) @protobuf(7,bytes,opt)

	// expirationSeconds is the requested duration of validity of the issued
	// certificate. The certificate signer may issue a certificate with a different
	// validity duration so a client must check the delta between the notBefore and
	// and notAfter fields in the issued certificate to determine the actual duration.
	//
	// The v1.22+ in-tree implementations of the well-known Kubernetes signers will
	// honor this field as long as the requested duration is not greater than the
	// maximum duration they will honor per the --cluster-signing-duration CLI
	// flag to the Kubernetes controller manager.
	//
	// Certificate signers may not honor this field for various reasons:
	//
	//   1. Old signer that is unaware of the field (such as the in-tree
	//      implementations prior to v1.22)
	//   2. Signer whose configured maximum is shorter than the requested duration
	//   3. Signer whose configured minimum is longer than the requested duration
	//
	// The minimum valid value for expirationSeconds is 600, i.e. 10 minutes.
	//
	// As of v1.22, this field is beta and is controlled via the CSRDuration feature gate.
	//
	// +optional
	expirationSeconds?: null | int32 @go(ExpirationSeconds,*int32) @protobuf(8,varint,opt)

	// allowedUsages specifies a set of usage contexts the key will be
	// valid for.
	// See: https://tools.ietf.org/html/rfc5280#section-4.2.1.3
	//      https://tools.ietf.org/html/rfc5280#section-4.2.1.12
	// Valid values are:
	//  "signing",
	//  "digital signature",
	//  "content commitment",
	//  "key encipherment",
	//  "key agreement",
	//  "data encipherment",
	//  "cert sign",
	//  "crl sign",
	//  "encipher only",
	//  "decipher only",
	//  "any",
	//  "server auth",
	//  "client auth",
	//  "code signing",
	//  "email protection",
	//  "s/mime",
	//  "ipsec end system",
	//  "ipsec tunnel",
	//  "ipsec user",
	//  "timestamping",
	//  "ocsp signing",
	//  "microsoft sgc",
	//  "netscape sgc"
	// +listType=atomic
	usages?: [...#KeyUsage] @go(Usages,[]KeyUsage) @protobuf(5,bytes,opt)

	// Information about the requesting user.
	// See user.Info interface for details.
	// +optional
	username?: string @go(Username) @protobuf(2,bytes,opt)

	// UID information about the requesting user.
	// See user.Info interface for details.
	// +optional
	uid?: string @go(UID) @protobuf(3,bytes,opt)

	// Group information about the requesting user.
	// See user.Info interface for details.
	// +listType=atomic
	// +optional
	groups?: [...string] @go(Groups,[]string) @protobuf(4,bytes,rep)

	// Extra information about the requesting user.
	// See user.Info interface for details.
	// +optional
	extra?: {[string]: #ExtraValue} @go(Extra,map[string]ExtraValue) @protobuf(6,bytes,rep)
}

// Signs certificates that will be honored as client-certs by the
// kube-apiserver. Never auto-approved by kube-controller-manager.
#KubeAPIServerClientSignerName: "kubernetes.io/kube-apiserver-client"

// Signs client certificates that will be honored as client-certs by the
// kube-apiserver for a kubelet.
// May be auto-approved by kube-controller-manager.
#KubeAPIServerClientKubeletSignerName: "kubernetes.io/kube-apiserver-client-kubelet"

// Signs serving certificates that are honored as a valid kubelet serving
// certificate by the kube-apiserver, but has no other guarantees.
#KubeletServingSignerName: "kubernetes.io/kubelet-serving"

// Has no guarantees for trust at all. Some distributions may honor these
// as client certs, but that behavior is not standard kubernetes behavior.
#LegacyUnknownSignerName: "kubernetes.io/legacy-unknown"

// ExtraValue masks the value so protobuf can generate
// +protobuf.nullable=true
// +protobuf.options.(gogoproto.goproto_stringer)=false
#ExtraValue: [...string]

#CertificateSigningRequestStatus: {
	// Conditions applied to the request, such as approval or denial.
	// +listType=map
	// +listMapKey=type
	// +optional
	conditions?: [...#CertificateSigningRequestCondition] @go(Conditions,[]CertificateSigningRequestCondition) @protobuf(1,bytes,rep)

	// If request was approved, the controller will place the issued certificate here.
	// +listType=atomic
	// +optional
	certificate?: bytes @go(Certificate,[]byte) @protobuf(2,bytes,opt)
}

#RequestConditionType: string // #enumRequestConditionType

#enumRequestConditionType:
	#CertificateApproved |
	#CertificateDenied |
	#CertificateFailed

#CertificateApproved: #RequestConditionType & "Approved"
#CertificateDenied:   #RequestConditionType & "Denied"
#CertificateFailed:   #RequestConditionType & "Failed"

#CertificateSigningRequestCondition: {
	// type of the condition. Known conditions include "Approved", "Denied", and "Failed".
	type: #RequestConditionType @go(Type) @protobuf(1,bytes,opt,casttype=RequestConditionType)

	// Status of the condition, one of True, False, Unknown.
	// Approved, Denied, and Failed conditions may not be "False" or "Unknown".
	// Defaults to "True".
	// If unset, should be treated as "True".
	// +optional
	status: v1.#ConditionStatus @go(Status) @protobuf(6,bytes,opt,casttype=k8s.io/api/core/v1.ConditionStatus)

	// brief reason for the request state
	// +optional
	reason?: string @go(Reason) @protobuf(2,bytes,opt)

	// human readable message with details about the request state
	// +optional
	message?: string @go(Message) @protobuf(3,bytes,opt)

	// timestamp for the last update to this condition
	// +optional
	lastUpdateTime?: metav1.#Time @go(LastUpdateTime) @protobuf(4,bytes,opt)

	// lastTransitionTime is the time the condition last transitioned from one status to another.
	// If unset, when a new condition type is added or an existing condition's status is changed,
	// the server defaults this to the current time.
	// +optional
	lastTransitionTime?: metav1.#Time @go(LastTransitionTime) @protobuf(5,bytes,opt)
}

#CertificateSigningRequestList: {
	metav1.#TypeMeta

	// +optional
	metadata?: metav1.#ListMeta @go(ListMeta) @protobuf(1,bytes,opt)
	items: [...#CertificateSigningRequest] @go(Items,[]CertificateSigningRequest) @protobuf(2,bytes,rep)
}

// KeyUsages specifies valid usage contexts for keys.
// See: https://tools.ietf.org/html/rfc5280#section-4.2.1.3
//      https://tools.ietf.org/html/rfc5280#section-4.2.1.12
#KeyUsage: string // #enumKeyUsage

#enumKeyUsage:
	#UsageSigning |
	#UsageDigitalSignature |
	#UsageContentCommitment |
	#UsageKeyEncipherment |
	#UsageKeyAgreement |
	#UsageDataEncipherment |
	#UsageCertSign |
	#UsageCRLSign |
	#UsageEncipherOnly |
	#UsageDecipherOnly |
	#UsageAny |
	#UsageServerAuth |
	#UsageClientAuth |
	#UsageCodeSigning |
	#UsageEmailProtection |
	#UsageSMIME |
	#UsageIPsecEndSystem |
	#UsageIPsecTunnel |
	#UsageIPsecUser |
	#UsageTimestamping |
	#UsageOCSPSigning |
	#UsageMicrosoftSGC |
	#UsageNetscapeSGC

#UsageSigning:           #KeyUsage & "signing"
#UsageDigitalSignature:  #KeyUsage & "digital signature"
#UsageContentCommitment: #KeyUsage & "content commitment"
#UsageKeyEncipherment:   #KeyUsage & "key encipherment"
#UsageKeyAgreement:      #KeyUsage & "key agreement"
#UsageDataEncipherment:  #KeyUsage & "data encipherment"
#UsageCertSign:          #KeyUsage & "cert sign"
#UsageCRLSign:           #KeyUsage & "crl sign"
#UsageEncipherOnly:      #KeyUsage & "encipher only"
#UsageDecipherOnly:      #KeyUsage & "decipher only"
#UsageAny:               #KeyUsage & "any"
#UsageServerAuth:        #KeyUsage & "server auth"
#UsageClientAuth:        #KeyUsage & "client auth"
#UsageCodeSigning:       #KeyUsage & "code signing"
#UsageEmailProtection:   #KeyUsage & "email protection"
#UsageSMIME:             #KeyUsage & "s/mime"
#UsageIPsecEndSystem:    #KeyUsage & "ipsec end system"
#UsageIPsecTunnel:       #KeyUsage & "ipsec tunnel"
#UsageIPsecUser:         #KeyUsage & "ipsec user"
#UsageTimestamping:      #KeyUsage & "timestamping"
#UsageOCSPSigning:       #KeyUsage & "ocsp signing"
#UsageMicrosoftSGC:      #KeyUsage & "microsoft sgc"
#UsageNetscapeSGC:       #KeyUsage & "netscape sgc"
