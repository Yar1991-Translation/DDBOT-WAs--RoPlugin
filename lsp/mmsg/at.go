package mmsg

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	localutils "github.com/cnxysoft/DDBOT-WSa/utils"
	"strconv"
)

type AtElement struct {
	*message.AtElement
}

func (a *AtElement) Type() message.ElementType {
	return At
}

func (a *AtElement) PackToElement(target Target) message.IMessageElement {
	if a == nil || a.AtElement == nil {
		return nil
	}
	switch target.TargetType() {
	case TargetGroup:
		if a.Target == 0 {
			a.QQ = "all"
			a.Display = "@全体成员"
		} else {
			a.QQ = strconv.FormatInt(a.Target, 10)
			if a.Display == "" {
				if gi := localutils.GetBot().FindGroup(target.TargetCode()); gi != nil {
					if gmi := gi.FindMember(a.Target); gmi != nil {
						a.Display = fmt.Sprintf("@%v", gmi.DisplayName())
					}
				}
			}
			if a.Display == "" {
				a.Display = fmt.Sprintf("@%v", a.Target)
			}
		}
		return a.AtElement
	default:
		return nil
	}
}

func NewAt(target int64, display ...string) *AtElement {
	var dis string
	if len(display) != 0 {
		dis = display[0]
	}
	return &AtElement{
		AtElement: &message.AtElement{
			Target:  target,
			Display: dis,
		},
	}
}
