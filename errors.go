package main

import "errors"

var (
	ErrNotFound     = errors.New("ressource non trouvée")
	ErrInvalidInput = errors.New("les données d'entrée sont invalides ou incomplètes")

	ErrSkillMissing = errors.New("l'utilisateur ne possède pas la compétence requise pour ce service")

	ErrSelfExchange         = errors.New("un utilisateur ne peut pas s'échanger un service à lui-même")
	ErrInsufficientCredits  = errors.New("l'utilisateur ne dispose pas d'un solde de crédits suffisant")
	ErrServiceUnavailable   = errors.New("ce service a déjà un échange en cours (statut pending ou accepted)")
	ErrInvalidTransition    = errors.New("changement de statut de l'échange non autorisé")

	ErrReviewNotAllowed = errors.New("impossible de noter un échange non terminé ou déjà évalué")
)
