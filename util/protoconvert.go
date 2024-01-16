package util

import (
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func ProtoStringToPG(s *wrapperspb.StringValue) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s.Value, Valid: true}
}

func PGToProtoString(s pgtype.Text) *wrapperspb.StringValue {
	if !s.Valid {
		return nil
	}
	return &wrapperspb.StringValue{Value: s.String}
}
