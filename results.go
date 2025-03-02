package mingodb

type InsertID interface{}

type SingleResult struct {
	data []byte
}

type MultiResult struct {
	//data []byte
	ResultCount int // Number of returned results
}

type UpdateResult struct {
	UpdateCount int // Number of rows updated
}

type DeleteResult struct {
	DeleteCount int // Number of rows deleted
}
