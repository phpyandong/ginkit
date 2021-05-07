package ging

import (
	"github.com/gin-gonic/gin"
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 控制器接口
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
type IController interface {
	Action(action func(ctx *gin.Context) IActionResult, args ...interface{}) func(ctx *gin.Context)
	SetCtrlFilters(filters ...IActionFilter) IController
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 控制器数据结构
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
type (
	Controller struct {
		filters []IActionFilter
	}
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 控制器动作
 * arg[0].(bool)==fasle关闭当前控制器的过滤器 | args many IActionFilter
 * 控制器执行流程：控制器Before->方法Before->方法After->控制器After
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (ctrl *Controller) Action(action func(ctx *gin.Context) IActionResult, args ...interface{}) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		actionFilters := make([]IActionFilter, 0)

		var filterResult IActionResult
		isEnabled := true
		argsCount := len(args)
		if argsCount > 0 {
			if value, isOk := args[0].(bool); isOk {
				isEnabled = value
			} else {
				for _, actionFilter := range args {
					if actionFilter, isOk := actionFilter.(IActionFilter); isOk {
						actionFilters = append(actionFilters, actionFilter)
					}
				}
			}
		}

		//启用拦截
		if isEnabled {
			//控制器的Before拦截器拦截
			for _, ctrlBeforeFilter := range ctrl.filters {
				if ctrlBeforeFilter != nil {
					if filterResult = ctrlBeforeFilter.Before(ctx); filterResult != nil {
						break
					}
				}

			}

			//控制的Before拦截器通过
			if filterResult == nil {
				//方法Before拦截
				for _, actionBeforeFilter := range actionFilters {
					if actionBeforeFilter != nil {
						if filterResult = actionBeforeFilter.Before(ctx); filterResult != nil {
							break
						}
					}
				}
			}
		}

		//拦截成功
		if filterResult != nil {
			filterResult.Render() //渲染拦截结果
		} else {
			action(ctx).Render() //执行真正的Handler方法，渲染返回结果
		}

		if isEnabled {
			//控制器的After过滤器
			for _, ctrlAfterFilter := range ctrl.filters {
				if ctrlAfterFilter != nil {
					ctrlAfterFilter.After(ctx) //一次执行过滤
				}
			}

			//方法的After过滤器
			for _, actionAfterFilter := range actionFilters {
				if actionAfterFilter != nil {
					actionAfterFilter.After(ctx)
				}
			}
		}
	}
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 设置控制器的拦截器
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (ctrl *Controller) SetCtrlFilters(filters ...IActionFilter) IController {
	if len(filters) == 0 {
		return ctrl
	}
	ctrl.filters = make([]IActionFilter, 0)
	for _, ctrlFilter := range filters {
		ctrl.filters = append(ctrl.filters, ctrlFilter)
	}
	return ctrl
}
