package main

func requireOwner(targetID, requesterID int) error {
	if requesterID != targetID {
		return ErrForbidden
	}
	return nil
}
