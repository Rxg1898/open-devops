package models

import (
	"fmt"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"open-devops/src/common"
)

type StreePath struct {
	Id       int64  `json:"id"`
	Level    int64  `json:"level"`
	Path     string `json:"path"`
	NodeName string `json:"node_name"`
}

// 插入一条记录
func (sp *StreePath) AddOne() (int64, error) {
	rowAffect, err := DB["stree"].InsertOne(sp)
	return rowAffect, err
}

// 根据部分条件查询一条记录
func (sp *StreePath) GetOne() (*StreePath, error) {
	has, err := DB["stree"].Get(sp)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return sp, nil
}

// 删除一条记录
func (sp *StreePath) DelOne() (int64, error) {
	return DB["stree"].Delete(sp)
}

// 检查一条记录是否存在
func (sp *StreePath) CheckExist() (bool, error) {
	return DB["stree"].Exist(sp)
}

// 函数区

// 增加一条path记录
func StreePathAddOne(req *common.NodeCommonReq, logger log.Logger) {
	// 要求新增的式 g.p.a 3段式
	res := strings.Split(req.Node, ".")
	if len(res) != 3 {
		level.Info(logger).Log("msg", "add.path.invalidate", "path", req.Node)
		return
	}

	// g.p.a
	g, p, a := res[0], res[1], res[2]

	// 先查g
	nodeG := &StreePath{
		Level:    1,
		Path:     "0",
		NodeName: g,
	}
	dbG, err := nodeG.GetOne()
	if err != nil {
		level.Error(logger).Log("msg", "check.g.failed", "path", req.Node, "err", err)
		return
	}
	// 根据g查询结果再判断
	switch dbG {
	case nil:
		// 说明g不存在，依次插入g.p.a
		// 插入g
		_, err := nodeG.AddOne()
		if err != nil {
			level.Error(logger).Log("msg", "g_not_exist_add_g_faild", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_not_exist_add_g_success", "path", req.Node)
		// 插入p
		pathP := fmt.Sprintf("/%d", nodeG.Id)
		nodeP := &StreePath{
			Level:    2,
			Path:     pathP,
			NodeName: p,
		}
		_, err = nodeP.AddOne()
		if err != nil {
			level.Error(logger).Log("msg", "g_not_exist_add_p_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_not_exist_add_p_success", "path", req.Node)

		// 插入a
		pathA := fmt.Sprintf("%s/%d", pathP, nodeP.Id)
		nodeA := &StreePath{
			Level:    3,
			Path:     pathA,
			NodeName: a,
		}
		_, err = nodeA.AddOne()
		if err != nil {
			level.Error(logger).Log("msg", "g_not_exist_add_a_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_not_exist_add_a_success", "path", req.Node)

	default:
		level.Info(logger).Log("msg", "g_exist_check_p", "path", req.Node)
		// 说明g存在，再查p
		pathP := fmt.Sprintf("/%d", dbG.Id)
		nodeP := &StreePath{
			Level:    2,
			Path:     pathP,
			NodeName: p,
		}
		dbP, err := nodeP.GetOne()
		if err != nil {
			level.Error(logger).Log("msg", "g_exist_check_p_failed", "path", req.Node, "err", err)
			return
		}
		if dbP != nil {
			// 说明p存在，继续查a
			level.Info(logger).Log("msg", "g_p_exist_check_a", "path", req.Node)
			pathA := fmt.Sprintf("%s/%d", pathP, dbP.Id)
			nodeA := &StreePath{
				Level:    3,
				Path:     pathA,
				NodeName: a,
			}
			dbA, err := nodeA.GetOne()
			if err != nil {
				level.Error(logger).Log("msg", "g_p_exist_check_a_failed", "path", req.Node, "err", err)
				return
			}
			if dbA == nil {
				// 说明a不存在，插入a
				_, err := nodeA.AddOne()
				if err != nil {
					level.Error(logger).Log("msg", "g_p_exist_add_a_failed", "path", req.Node, "err", err)
					return
				}
				level.Info(logger).Log("msg", "g_p_exist_add_a_success", "path", req.Node)
				return
			}
			level.Info(logger).Log("msg", "g_p_a_exist", "path", req.Node)
			return
		}
		// 说明p不存在，插入p和a
		level.Info(logger).Log("msg", "g_exist_p_a_not", "path", req.Node)
		_, err = nodeP.AddOne()
		if err != nil {
			level.Error(logger).Log("msg", "g_exist_add_p_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_exist_add_p_success", "path", req.Node)
		// 插入a
		pathA := fmt.Sprintf("%s/%d", pathP, nodeP.Id)
		nodeA := &StreePath{
			Level:    3,
			Path:     pathA,
			NodeName: a,
		}
		_, err = nodeA.AddOne()
		if err != nil {
			level.Error(logger).Log("msg", "g_exist_add_a_failed", "path", req.Node, "err", err)
			return
		}
		level.Info(logger).Log("msg", "g_exist_add_a_success", "path", req.Node)

	}
}

// 编写新增node的测试函数
func StreePathAddTest(logger log.Logger) {
	ns := []string{
		"inf.monitor.thanos",
		"inf.monitor.kafka",
		"inf.monitor.prometheus",
		"inf.monitor.m3db",
	}
	for _, n := range ns {
		req := &common.NodeCommonReq{
			Node: n,
		}
		StreePathAddOne(req, logger)
	}
}
