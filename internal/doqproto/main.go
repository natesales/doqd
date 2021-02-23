package doqproto

// Only implementations of the final, published RFC can identify
// themselves as "doq". Until such an RFC exists, implementations MUST
// NOT identify themselves using this string. Implementations of draft
// versions of the protocol MUST add the string "-" and the corresponding
// draft number to the identifier. For example, draft-ietf-dprive-dnsoquic-00
// is identified using the string "doq-i00".

// TlsProtos stores the dnsoquic draft version for TLS protocol announcement
var TlsProtos = []string{"doq-i02"}

// TlsProtosCompat stores alternative TLS protocols for experimental interoperability
var TlsProtosCompat = []string{"doq-i02", "doq", "doq"}
