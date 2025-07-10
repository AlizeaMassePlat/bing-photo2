package main

import (
	"bytes"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"


	"GalleryService/internal/db"
	"GalleryService/internal/jwt"
	"GalleryService/internal/middleware"
	"GalleryService/internal/models"
	proto "GalleryService/internal/proto"
	"GalleryService/internal/services"
	"GalleryService/internal/utils"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

type galleryServer struct {
	proto.UnimplementedAlbumServiceServer
	proto.UnimplementedMediaServiceServer
	proto.UnimplementedUserServiceServer
	proto.UnimplementedConsentServiceServer
	albumService   *services.AlbumService
	mediaService   *services.MediaService
	userService    *services.UserService
	consentService *services.ConsentService
}

// Album Service methods
func (s *galleryServer) CreateAlbum(ctx context.Context, req *proto.CreateAlbumRequest) (*proto.CreateAlbumResponse, error) {
	album := models.Album{
		Name:        req.Name,
		Description: req.Description,
		UserID:      uint(req.UserId),
	}

	if err := s.albumService.CreateAlbum(album); err != nil {
		log.Printf("Error creating album: %v", err)
		return nil, err
	}

	return &proto.CreateAlbumResponse{
		Message: "Album created successfully",
	}, nil
}

func (s *galleryServer) GetAlbumsByUser(ctx context.Context, req *proto.GetAlbumsByUserRequest) (*proto.GetAlbumsByUserResponse, error) {
	albums, err := s.albumService.GetAlbumsByUser(uint(req.UserId))
	if err != nil {
		log.Printf("Error getting albums by user: %v", err)
		return nil, err
	}

	var protoAlbums []*proto.AlbumWithMedia
	for _, album := range albums {
		var protoMedia []*proto.Media
		for _, media := range album.Media {
			protoMedia = append(protoMedia, &proto.Media{
				Id:       uint32(media.ID),
				Name:     media.Name,
				Path:     media.Path,
				FileSize: uint32(media.FileSize),
				AlbumId:  uint32(media.AlbumID),
			})
		}

		protoAlbums = append(protoAlbums, &proto.AlbumWithMedia{
			Id:          uint32(album.ID),
			Name:        album.Name,
			Description: album.Description,
			UserId:      uint32(album.UserID),
			Media:       protoMedia,
		})
	}

	return &proto.GetAlbumsByUserResponse{
		Albums: protoAlbums,
	}, nil
}



func (s *galleryServer) UpdateAlbum(ctx context.Context, req *proto.UpdateAlbumRequest) (*proto.UpdateAlbumResponse, error) {
	if err := s.albumService.UpdateAlbum(uint(req.AlbumId), req.Name, req.Description); err != nil {
		log.Printf("Error updating album: %v", err)
		return nil, err
	}

	return &proto.UpdateAlbumResponse{}, nil
}

func (s *galleryServer) DeleteAlbum(ctx context.Context, req *proto.DeleteAlbumRequest) (*proto.DeleteAlbumResponse, error) {
	if err := s.albumService.DeleteAlbum(uint(req.AlbumId)); err != nil {
		log.Printf("Error deleting album: %v", err)
		return nil, err
	}

	return &proto.DeleteAlbumResponse{}, nil
}

func (s *galleryServer) GetPrivateAlbum(ctx context.Context, req *proto.GetPrivateAlbumRequest) (*proto.GetPrivateAlbumResponse, error) {
	album, err := s.albumService.GetPrivateAlbum(uint(req.UserId), req.Type)
	if err != nil {
		log.Printf("Error getting private album: %v", err)
		return nil, err
	}

	return &proto.GetPrivateAlbumResponse{
		Album: &proto.Album{
			Id:          uint32(album.ID),
			Name:        album.Name,
			Description: album.Description,
			UserId:      uint32(album.UserID),
		},
	}, nil
}

// Media Service methods
func (s *galleryServer) AddMedia(ctx context.Context, req *proto.AddMediaRequest) (*proto.AddMediaResponse, error) {
	media := &models.Media{
		Name:    req.Name,
		AlbumID: uint(req.AlbumId),
	}

	reader := bytes.NewReader(req.FileData)
	if err := s.mediaService.AddMedia(media, reader, int64(len(req.FileData))); err != nil {
		log.Printf("Error adding media: %v", err)
		return nil, err
	}

	return &proto.AddMediaResponse{
		Message: "Media added successfully",
	}, nil
}

func (s *galleryServer) GetMediaByUser(ctx context.Context, req *proto.GetMediaByUserRequest) (*proto.GetMediaByUserResponse, error) {
	media, err := s.mediaService.GetMediaByUser(uint(req.UserId))
	if err != nil {
		log.Printf("Error getting media by user: %v", err)
		return nil, err
	}

	var protoMedia []*proto.Media
	for _, m := range media {
		protoMedia = append(protoMedia, &proto.Media{
			Id:       uint32(m.ID),
			Name:     m.Name,
			AlbumId:  uint32(m.AlbumID),
			FileSize: uint32(m.FileSize),
		})
	}

	return &proto.GetMediaByUserResponse{
		MediaList: protoMedia,
	}, nil
}

func (s *galleryServer) MarkAsPrivate(ctx context.Context, req *proto.MarkAsPrivateRequest) (*proto.MarkAsPrivateResponse, error) {
	// Extraire l'identifiant utilisateur depuis le contexte
	userID, err := jwt.ExtractUserIDFromContext(ctx)
	if err != nil {
		log.Printf("[MarkAsPrivate] Échec extraction userID : %v", err)
		return nil, status.Error(codes.Unauthenticated, "Utilisateur non authentifié")
	}

	// Appeler le service avec le paramètre simulate (utile pour les tests ou la vérification avant création)
	log.Printf("[MarkAsPrivate] Appel service avec simulate=%v, media_id=%d", req.Simulate, req.MediaId)
	pinRequired, err := s.mediaService.MarkAsPrivate(uint(req.MediaId), userID, req.Simulate)
	if err != nil {
		log.Printf("[MarkAsPrivate] Erreur service : %v", err)
		return nil, status.Error(codes.Internal, "Erreur lors du marquage du média comme privé")
	}

	return &proto.MarkAsPrivateResponse{
		Message:     "Média marqué comme privé avec succès",
		PinRequired: pinRequired,
	}, nil
}
func (s *galleryServer) GetPrivateMedia(ctx context.Context, req *proto.GetPrivateMediaRequest) (*proto.GetPrivateMediaResponse, error) {
	// Extraire le userID depuis le contexte JWT
	userID, err := jwt.ExtractUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Impossible d'extraire le userID : %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou manquant")
	}

	// Récupérer les médias privés de l'utilisateur
	medias, err := s.mediaService.GetPrivateMedia(userID)
	if err != nil {
		log.Printf("Error getting private media: %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de la récupération des médias privés: %v", err)
	}

	// Convertir en proto
	var protoMedias []*proto.Media
	for _, media := range medias {
		protoMedia := &proto.Media{
			Id:        uint32(media.ID),
			Name:      media.Name,
			AlbumId:   uint32(media.AlbumID),
			FileSize:  uint32(media.FileSize),
			Path:      media.Path,
			IsPrivate: media.IsPrivate,
			IsFavorite: media.IsFavorite,
		}
		protoMedias = append(protoMedias, protoMedia)
	}

	return &proto.GetPrivateMediaResponse{
		Media: protoMedias,
	}, nil
}

func (s *galleryServer) DownloadMedia(ctx context.Context, req *proto.DownloadMediaRequest) (*proto.DownloadMediaResponse, error) {
	userID, err := jwt.ExtractUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Erreur d'extraction du userID : %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "token invalide : %v", err)
	}

	log.Printf("Téléchargement demandé pour mediaID=%d par userID=%d", req.MediaId, userID)

	var buf bytes.Buffer
	if err := s.mediaService.DownloadMedia(uint(req.MediaId), userID, &buf); err != nil {
		log.Printf("Erreur lors du téléchargement du média : %v", err)
		return nil, status.Errorf(codes.Internal, "échec du téléchargement du média : %v", err)
	}

	return &proto.DownloadMediaResponse{
		FileData: buf.Bytes(),
	}, nil
}

func (s *galleryServer) DeleteMedia(ctx context.Context, req *proto.DeleteMediaRequest) (*proto.DeleteMediaResponse, error) {

	userID, err := jwt.ExtractUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Erreur d'extraction du userID : %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "token invalide : %v", err)
	}

	if err := s.mediaService.DeleteMedia(uint(req.MediaId), userID); err != nil {
		log.Printf("Erreur lors de la suppression du média : %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de la suppression du média : %v", err)
	}

	return &proto.DeleteMediaResponse{
		Message: "Média supprimé avec succès",
	}, nil
}

func (s *galleryServer) DetectSimilarMedia(ctx context.Context, req *proto.DetectSimilarMediaRequest) (*proto.DetectSimilarMediaResponse, error) {
	userID, err := jwt.ExtractUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Token invalide : %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "token invalide : %v", err)
	}

	log.Printf("🔍 Détection de similarité sur albumID=%d pour userID=%d", req.AlbumId, userID)

	// Appel à la méthode modifiée
	similarGroups, err := s.mediaService.DetectSimilarMedia(userID, uint(req.AlbumId))
	if err != nil {
		log.Printf("Erreur détection similarité : %v", err)
		return nil, status.Errorf(codes.Unknown, err.Error())
	}

	// Convertit les groupes en format gRPC
	var protoGroups []*proto.MediaGroup
	for _, group := range similarGroups {
		var protoMedia []*proto.Media
		for _, m := range group {
			protoMedia = append(protoMedia, &proto.Media{
				Id:       uint32(m.ID),
				Name:     m.Name,
				AlbumId:  uint32(m.AlbumID),
				FileSize: uint32(m.FileSize),
				Path:     m.Path,
			})
		}
		protoGroups = append(protoGroups, &proto.MediaGroup{Media: protoMedia})
	}

	return &proto.DetectSimilarMediaResponse{
		Groups: protoGroups,
	}, nil
}

// User Service methods
func (s *galleryServer) CreateUser(ctx context.Context, req *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	if err := s.userService.CreateUser(req.Username, req.Email); err != nil {
		log.Printf("Error creating user: %v", err)
		return nil, err
	}

	return &proto.CreateUserResponse{}, nil
}

// func (s *galleryServer) AddMediaToFavorite(ctx context.Context, req *proto.AddMediaToFavoriteRequest) (*proto.AddMediaToFavoriteResponse, error) {
// 	if err := s.mediaService.AddMediaToFavorite(uint(req.MediaId)); err != nil {
// 		log.Printf("Error adding media to favorite: %v", err)
// 		return nil, err
// 	}

// 	return &proto.AddMediaToFavoriteResponse{}, nil
// }

func (s *galleryServer) GetMediaByAlbum(ctx context.Context, req *proto.GetMediaByAlbumRequest) (*proto.GetMediaByAlbumResponse, error) {
    medias, err := s.mediaService.GetMediaByAlbum(uint(req.AlbumId))
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to retrieve media: %v", err)
    }

    var protoMedias []*proto.Media
    for _, m := range medias {
        protoMedias = append(protoMedias, &proto.Media{
            Id:        uint32(m.ID),
            Name:      m.Name,
            Path:      m.Path,
            AlbumId:   uint32(m.AlbumID),
			IsFavorite: m.IsFavorite,
        })
    }

    return &proto.GetMediaByAlbumResponse{Media: protoMedias}, nil
}

func (s *galleryServer) MoveMediaToAlbum(ctx context.Context, req *proto.MoveMediaRequest) (*proto.MoveMediaResponse, error) {
	if err := s.mediaService.MoveMediaToAlbum(uint(req.MediaId), uint(req.TargetAlbumId)); err != nil {
		log.Printf("Erreur lors du déplacement du média : %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors du déplacement du média : %v", err)
	}
	return &proto.MoveMediaResponse{Message: "Média déplacé avec succès"}, nil
}

// Consent Service methods
func (s *galleryServer) AddConsent(ctx context.Context, req *proto.AddConsentRequest) (*proto.AddConsentResponse, error) {
	// Extraire le userID depuis le contexte
	userID, err := jwt.ExtractUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Impossible d'extraire le userID : %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou manquant")
	}

	consentReq := &services.ConsentRequest{
		ConsentType:    req.ConsentType,
		ConsentVersion: req.ConsentVersion,
		IsGranted:      req.IsGranted,
		ConsentText:    req.ConsentText,
	}

	// Créer une requête HTTP factice pour récupérer l'IP et User-Agent
	// En gRPC, nous devons les passer via les métadonnées ou les paramètres
	consent, err := s.consentService.AddConsent(userID, consentReq, nil)
	if err != nil {
		log.Printf("Error adding consent: %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de l'ajout du consentement: %v", err)
	}

	protoConsent := &proto.Consent{
		Id:             uint32(consent.ID),
		ConsentType:    consent.ConsentType,
		ConsentVersion: consent.ConsentVersion,
		IsGranted:      consent.IsGranted,
		ConsentText:    consent.ConsentText,
		RevokedAt:      "",
		CreatedAt:      consent.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      consent.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if consent.RevokedAt != nil {
		protoConsent.RevokedAt = consent.RevokedAt.Format("2006-01-02T15:04:05Z")
	}

	return &proto.AddConsentResponse{
		Message:  "Consentement ajouté avec succès",
		Consent:  protoConsent,
	}, nil
}

func (s *galleryServer) GetUserConsents(ctx context.Context, req *proto.GetUserConsentsRequest) (*proto.GetUserConsentsResponse, error) {
	// Extraire le userID depuis le contexte
	userID, err := jwt.ExtractUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Impossible d'extraire le userID : %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou manquant")
	}

	consents, err := s.consentService.GetUserConsents(userID)
	if err != nil {
		log.Printf("Error getting user consents: %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de la récupération des consentements: %v", err)
	}

	var protoConsents []*proto.Consent
	for _, consent := range consents {
		protoConsent := &proto.Consent{
			Id:             uint32(consent.ID),
			ConsentType:    consent.ConsentType,
			ConsentVersion: consent.ConsentVersion,
			IsGranted:      consent.IsGranted,
			ConsentText:    consent.ConsentText,
			RevokedAt:      "",
			CreatedAt:      consent.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      consent.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if consent.RevokedAt != nil {
			protoConsent.RevokedAt = consent.RevokedAt.Format("2006-01-02T15:04:05Z")
		}

		protoConsents = append(protoConsents, protoConsent)
	}

	return &proto.GetUserConsentsResponse{
		Consents: protoConsents,
		Count:    uint32(len(protoConsents)),
	}, nil
}

func (s *galleryServer) GetActiveConsent(ctx context.Context, req *proto.GetActiveConsentRequest) (*proto.GetActiveConsentResponse, error) {
	// Extraire le userID depuis le contexte
	userID, err := jwt.ExtractUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Impossible d'extraire le userID : %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou manquant")
	}

	consent, err := s.consentService.GetActiveConsent(userID, req.ConsentType)
	if err != nil {
		log.Printf("Error getting active consent: %v", err)
		return nil, status.Errorf(codes.NotFound, "Aucun consentement actif trouvé: %v", err)
	}

	protoConsent := &proto.Consent{
		Id:             uint32(consent.ID),
		ConsentType:    consent.ConsentType,
		ConsentVersion: consent.ConsentVersion,
		IsGranted:      consent.IsGranted,
		ConsentText:    consent.ConsentText,
		RevokedAt:      "",
		CreatedAt:      consent.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      consent.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if consent.RevokedAt != nil {
		protoConsent.RevokedAt = consent.RevokedAt.Format("2006-01-02T15:04:05Z")
	}

	return &proto.GetActiveConsentResponse{
		Consent: protoConsent,
	}, nil
}

func (s *galleryServer) CheckConsent(ctx context.Context, req *proto.CheckConsentRequest) (*proto.CheckConsentResponse, error) {
	// Extraire le userID depuis le contexte
	userID, err := jwt.ExtractUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Impossible d'extraire le userID : %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou manquant")
	}

	hasConsent, err := s.consentService.HasValidConsent(userID, req.ConsentType)
	if err != nil {
		log.Printf("Error checking consent: %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de la vérification du consentement: %v", err)
	}

	return &proto.CheckConsentResponse{
		HasConsent:   hasConsent,
		ConsentType:  req.ConsentType,
	}, nil
}

func (s *galleryServer) RevokeConsent(ctx context.Context, req *proto.RevokeConsentRequest) (*proto.RevokeConsentResponse, error) {
	// Extraire le userID depuis le contexte
	userID, err := jwt.ExtractUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Impossible d'extraire le userID : %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou manquant")
	}

	err = s.consentService.RevokeConsent(userID, uint(req.ConsentId), nil)
	if err != nil {
		log.Printf("Error revoking consent: %v", err)
		return nil, status.Errorf(codes.Internal, "Erreur lors de la révocation du consentement: %v", err)
	}

	return &proto.RevokeConsentResponse{
		Message: "Consentement révoqué avec succès",
	}, nil
}

func (s *galleryServer) SetUserPin(ctx context.Context, req *proto.SetUserPinRequest) (*proto.SetUserPinResponse, error) {
	userID := req.UserId
	if userID == 0 {
		var id uint
		var err error
		id, err = jwt.ExtractUserIDFromContext(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou manquant")
		}
		userID = uint32(id)
	}

	if len(req.Pin) < 4 {
		return nil, status.Errorf(codes.InvalidArgument, "Le PIN doit contenir au moins 4 chiffres")
	}

	// Utilise utils.HashPin
	hashedPin := utils.HashPin(req.Pin)

	err := s.userService.UpdatePin(uint(userID), hashedPin)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Erreur lors de la mise à jour du PIN : %v", err)
	}

	return &proto.SetUserPinResponse{
		Message: "PIN mis à jour avec succès",
	}, nil
}

func (s *galleryServer) VerifyUserPin(ctx context.Context, req *proto.VerifyUserPinRequest) (*proto.VerifyUserPinResponse, error) {
    userID := req.UserId
    if userID == 0 {
        var id uint
        var err error
        id, err = jwt.ExtractUserIDFromContext(ctx)
        if err != nil {
            log.Printf("Erreur extraction userID: %v", err) // <-- AJOUT
            return nil, status.Errorf(codes.Unauthenticated, "Token invalide ou manquant")
        }
        userID = uint32(id)
    }

    if req.Pin == "" {
        log.Printf("PIN vide reçu pour userID=%d", userID) // <-- AJOUT
        return nil, status.Errorf(codes.InvalidArgument, "Le PIN ne peut pas être vide")
    }

    // Vérifier le PIN
    isValid, err := s.userService.VerifyUserPin(uint(userID), req.Pin)
    if err != nil {
        log.Printf("Erreur service VerifyUserPin (userID=%d): %v", userID, err) // <-- AJOUT
        return nil, status.Errorf(codes.Internal, "Erreur lors de la vérification du PIN : %v", err)
    }

    log.Printf("Résultat vérification PIN userID=%d: %v", userID, isValid) // <-- AJOUT
    if isValid {
        return &proto.VerifyUserPinResponse{
            Valid:   true,
            Message: "PIN vérifié avec succès",
        }, nil
    } else {
        return &proto.VerifyUserPinResponse{
            Valid:   false,
            Message: "PIN incorrect",
        }, nil
    }
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Avertissement : Impossible de charger le fichier .env, utilisation des variables système.")
	}

	// Initialiser le gestionnaire de base de données
	dbManager, err := db.NewDBManagerService()
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de la base de données : %v", err)
	}
	defer func() {
		log.Println("Fermeture de la connexion à la base de données...")
		dbManager.CloseConnection()
	}()

	// Effectuer la migration des modèles
	if err := dbManager.AutoMigrate(); err != nil {
		log.Fatalf("Erreur lors de la migration des modèles : %v", err)
	}

	// Initialiser le S3Service
	s3Service := services.NewS3Service("http://my-s3-clone:9090")

	// Initialize services
	albumService := services.NewAlbumService(dbManager, s3Service)
	consentService := services.NewConsentService(dbManager)
	mediaService := services.NewMediaService(dbManager, s3Service, consentService)
	userService := services.NewUserService(dbManager, s3Service)

	// Initialiser le service JWT
	jwtService, err := jwt.NewJWTService()
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de JWTService : %v", err)
	}

	// Définir les méthodes protégées (authentification requise)
	methodsToIntercept := map[string]bool{
		"/proto.AlbumService/CreateAlbum":     true,
		"/proto.AlbumService/UpdateAlbum":     true,
		"/proto.AlbumService/DeleteAlbum":     true,
		"/proto.AlbumService/GetPrivateAlbum": true,
		"/proto.MediaService/AddMedia":        true,
		"/proto.MediaService/MarkAsPrivate":   true,
		"/proto.MediaService/GetPrivateMedia": true,
		"/proto.MediaService/DownloadMedia":   true,
		"/proto.MediaService/DeleteMedia":     true,
		"/proto.MediaService/GetMediaByAlbum": true,
		"/proto.MediaService/MoveMediaToAlbum": true,
		"/proto.ConsentService/AddConsent":    true,
		"/proto.ConsentService/GetUserConsents": true,
		"/proto.ConsentService/GetActiveConsent": true,
		"/proto.ConsentService/CheckConsent":  true,
		"/proto.ConsentService/RevokeConsent": true,
		"/proto.UserService/SetUserPin":        true,
		"/proto.UserService/VerifyUserPin":    true,
	}

	// Créer le serveur gRPC avec intercepteur JWT
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.AuthInterceptor(jwtService, methodsToIntercept)),
	)

	galleryServer := &galleryServer{
		albumService:   albumService,
		mediaService:   mediaService,
		userService:    userService,
		consentService: consentService,
	}

	// Enregistrer les services gRPC
	proto.RegisterAlbumServiceServer(grpcServer, galleryServer)
	proto.RegisterMediaServiceServer(grpcServer, galleryServer)
	proto.RegisterUserServiceServer(grpcServer, galleryServer)
	proto.RegisterConsentServiceServer(grpcServer, galleryServer)

	// Démarrer le serveur gRPC
	grpcListener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Erreur lors de l'écoute du serveur gRPC : %v", err)
	}

	// Canal pour gérer les signaux système (interruption ou arrêt)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Démarrer le serveur gRPC dans une goroutine
	go func() {
		log.Println("gRPC server started on port 50052...")
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Erreur lors de l'exécution du serveur gRPC : %v", err)
		}
	}()

	// Attendre un signal d'arrêt
	<-stop
	log.Println("Signal reçu, arrêt des services...")

	// Arrêter gracieusement les serveurs
	grpcServer.GracefulStop()
	log.Println("Server stopped successfully.")
}
