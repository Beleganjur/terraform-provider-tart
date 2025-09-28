package main

import "net/http"

// In production: implement mTLS, Unix socket, or JWT secure peer verification here

func isAuthorizedRequest(r *http.Request) bool {
    // Dummy: Always authorize
    return true
}
