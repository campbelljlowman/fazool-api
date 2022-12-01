package voter

import (
	"time"

)

type Voter struct {

}

func (v *model.Voter) refresh() {
	v.Expires = time.Now()
}