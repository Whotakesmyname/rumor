package service

// K sets K-bucket size
const K int = 8

// Port set the port used for listening
// Note: This option only defines local listen port. For terminals behind NAT(s), it will differ from local node's address.
const Port int = 54321

// NodeIDLength sets NodeID length in bytes. Default 20 in SHA-1.
const NodeIDLength int = 20

// CookieLength sets Magic Cookie length in bytes. Default 20 in SHA-1.
const CookieLength int = 20

// MaxTextLength sets Max Text length in UTF-8. Default 200.
const MaxTextLength int = 200

// MaxPackageSize sets Max UDP package size in bytes. Default 1460, considering PPPOE.
const MaxPackageSize int = 1460

// RequestTimeout sets Timeout of every request in seconds. Note: This is the least time a cookie would be preserved.
const RequestTimeout float64 = 60

// RefreshInternal sets Frequency of CookieTable Refresh. It's an interval in seconds.
const RefreshInternal int = 30

// ResponseHandlerQueueLength sets Response handler queue length
const ResponseHandlerQueueLength int = 16

// RequestHandlerQueueLength sets Response handler queue length
const RequestHandlerQueueLength int = 16
