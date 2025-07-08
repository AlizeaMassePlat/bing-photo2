package middleware

import (
    "net/http"
    "strings"
    "log"
    "bytes"
)
// CorsMiddleware permet de configurer les en-têtes CORS
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Autoriser les origines spécifiques
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
        
        // Autoriser les méthodes HTTP
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        
        // Autoriser les en-têtes spécifiques
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Amz-Content-Sha256, X-Amz-Decoded-Content-Length")
        
        // Gérer les requêtes preflight (OPTIONS)
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func BasicAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasPrefix(r.URL.Path, "/probe-bsign") {
            next.ServeHTTP(w, r)
            return
        }

        // Accepter les requêtes avec AWS4-HMAC-SHA256 dans l'en-tête Authorization
        if strings.Contains(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256") {
            next.ServeHTTP(w, r)
            return
        }

        // Appliquer l'authentification basique pour les autres routes
        user, pass, ok := r.BasicAuth()
        if !ok || user != "accessuser" || pass != "accesspassword" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func LogRequestMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Requête reçue 1: %s %s", r.Method, r.RequestURI)

        if len(r.URL.Query()) > 0 {
            log.Printf("Query Params: %v", r.URL.Query())
        }

        next.ServeHTTP(w, r)
    })
}

type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
    responseBody bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
    lrw.statusCode = code
    lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
    lrw.responseBody.Write(b) 
    return lrw.ResponseWriter.Write(b) 
}

func LogResponseMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        next.ServeHTTP(lrw, r)

        // Log la réponse
        log.Printf("Response status: %d", lrw.statusCode)

        contentType := w.Header().Get("Content-Type")

        // Log uniquement les types texte ou XML lisibles
        if strings.HasPrefix(contentType, "text/") || strings.Contains(contentType, "xml") || strings.Contains(contentType, "json") {
            // Tronque les réponses trop longues pour éviter les floods
            body := lrw.responseBody.String()
            maxLen := 5000
            if len(body) > maxLen {
                body = body[:maxLen] + "...(truncated)"
            }
            log.Printf("Response body: %s", body)
        } else {
            log.Printf("Response body not logged (binary content: %s)", contentType)
        }
    })
}
