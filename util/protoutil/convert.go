package protoutil

import (
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func WrappedStringToPGText(s *wrapperspb.StringValue) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s.Value, Valid: true}
}

func PGTextToWrappedString(s pgtype.Text) *wrapperspb.StringValue {
	if !s.Valid {
		return nil
	}
	return &wrapperspb.StringValue{Value: s.String}
}

func TimestampToPGTimestamptz(ts *timestamppb.Timestamp) pgtype.Timestamptz {
	if ts == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{
		Time:  ts.AsTime(),
		Valid: ts.IsValid(),
	}
}

func PGTimestamptzToTimestamp(ts pgtype.Timestamptz) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}
	return timestamppb.New(ts.Time)
}
