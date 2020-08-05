package notice

import (
	"context"
	"database/sql"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

// A Store allows identifying notices for various targets.
type Store struct {
	findServicesByPolicyID *sql.Stmt
}

// NewStore creates a new DB and prepares all necessary SQL statements.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &Store{
		findServicesByPolicyID: p.P(`
			SELECT COUNT(*)
			FROM services
			WHERE escalation_policy_id = $1
		`),
	}, p.Err
}

// Sets a notice for a Policy if it is not assigned to any services
func (s *Store) FindAllPolicyNotices(ctx context.Context, policyID string) ([]Notice, error) {
	err := validate.UUID("EscalationPolicyStepID", policyID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	var numServices int
	err = s.findServicesByPolicyID.QueryRowContext(ctx, policyID).Scan(&numServices)
	if err != nil {
		return nil, err
	}

	var notices = make([]Notice, 1)
	if numServices == 0 {
		notices[0].Type = Warning
		notices[0].Message = "Not assigned to a service"
		notices[0].Details = "To receive alerts for this configuration, assign this escalation policy to its relavent service."
		return notices, nil
	}

	// no results
	return nil, nil
}