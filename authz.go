package main

// requireOwner vérifie que requesterID est bien le propriétaire de la
// ressource targetID. Utilisé par tous les endpoints de modification qui
// s'appuient sur l'authentification "simple" par header X-User-ID.
func requireOwner(targetID, requesterID int) error {
	if requesterID != targetID {
		return ErrForbidden
	}
	return nil
}
