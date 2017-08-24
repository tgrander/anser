package anser

import (
	"github.com/mongodb/amboy"
	"github.com/mongodb/amboy/job"
	"github.com/mongodb/amboy/registry"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	registry.AddJobType("simple-migration",
		func() amboy.Job { return makeSimpleMigration() })
}

func NewSimpleMigration(e Environment, m SimpleMigration) amboy.Job {
	j := makeSimpleMigration()
	j.Definition = m
	j.MigrationHelper = NewMigrationHelper(e)
	return j
}

func makeSimpleMigration() *simpleMigrationJob {
	return &simpleMigrationJob{
		MigrationHelper: &migrationBase{},
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    "simple-migration",
				Version: 0,
				Format:  amboy.BSON,
			},
		},
	}
}

type simpleMigrationJob struct {
	Definition      SimpleMigration `bson:"migration" json:"migration" yaml:"migration"`
	job.Base        `bson:"job_base" json:"job_base" yaml:"job_base"`
	MigrationHelper `bson:"-" json:"-" yaml:"-"`
}

func (j *simpleMigrationJob) Run() {
	defer j.FinishMigration(j.Definition.Migration, &j.Base)

	env := j.Env()
	session, err := env.GetSession()
	if err != nil {
		j.AddError(errors.Wrap(err, "problem getting database session"))
		return
	}
	defer session.Close()

	coll := session.DB(j.Definition.Namespace.DB).C(j.Definition.Namespace.Collection)

	j.AddError(coll.UpdateId(j.Definition.ID, bson.M(j.Definition.Update)))
}