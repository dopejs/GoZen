package bot

// BuiltinSkills returns the default builtin skill definitions.
func BuiltinSkills() []*Skill {
	return []*Skill{
		{
			Name:        "process-control",
			Description: "Control process lifecycle: pause, resume, stop, cancel",
			Intent:      IntentControl,
			Priority:    10,
			Enabled:     true,
			Builtin:     true,
			Keywords: map[string][]string{
				"en": {"pause", "resume", "stop", "cancel", "start", "restart"},
				"zh": {"暂停", "恢复", "停止", "取消", "启动", "重启"},
			},
			Synonyms: map[string]string{
				"halt": "stop", "kill": "stop", "suspend": "pause",
				"continue": "resume", "unpause": "resume", "abort": "cancel",
				"中止": "停止", "挂起": "暂停", "继续": "恢复",
				"终止": "停止", "关掉": "停止", "停掉": "停止",
				"暂时停": "暂停", "先停": "暂停",
			},
			Examples: []string{"pause myapp", "帮我暂停一下", "stop the server"},
		},
		{
			Name:        "project-bind",
			Description: "Bind current chat to a project or target",
			Intent:      IntentBind,
			Priority:    20,
			Enabled:     true,
			Builtin:     true,
			Keywords: map[string][]string{
				"en": {"bind", "attach", "connect", "link"},
				"zh": {"绑定", "关联", "连接", "挂载"},
			},
			Synonyms: map[string]string{
				"bindto": "bind", "bindwith": "bind",
				"绑到": "绑定", "关联到": "关联", "对接": "连接",
			},
			Examples: []string{"bind myproject", "绑定到myproject"},
		},
		{
			Name:        "approval",
			Description: "Approve or reject pending requests",
			Intent:      IntentApprove,
			Priority:    30,
			Enabled:     true,
			Builtin:     true,
			Keywords: map[string][]string{
				"en": {"approve", "reject", "yes", "no", "deny", "accept"},
				"zh": {"批准", "拒绝", "同意", "否决", "通过", "驳回"},
			},
			Synonyms: map[string]string{
				"ok": "approve", "lgtm": "approve", "confirm": "approve",
				"decline": "reject", "nope": "reject",
				"好的": "批准", "不行": "拒绝", "可以": "批准",
				"行": "批准", "不同意": "拒绝", "没问题": "批准",
			},
			Examples: []string{"approve", "reject the request", "批准"},
		},
		{
			Name:        "send-task",
			Description: "Send a task to a specific target process",
			Intent:      IntentSendTask,
			Priority:    40,
			Enabled:     true,
			Builtin:     true,
			Keywords: map[string][]string{
				"en": {"send", "dispatch", "assign", "forward", "tell"},
				"zh": {"发送", "派发", "分配", "转发", "告诉"},
			},
			Synonyms: map[string]string{
				"sendto": "send", "push": "send",
				"发给": "发送", "交给": "分配", "传给": "转发",
			},
			Examples: []string{"send worker1 build the project", "告诉worker1编译项目"},
		},
		{
			Name:        "persona",
			Description: "Set, show, or clear bot persona",
			Intent:      IntentPersona,
			Priority:    50,
			Enabled:     true,
			Builtin:     true,
			Keywords: map[string][]string{
				"en": {"persona", "character", "role", "personality"},
				"zh": {"人设", "角色", "性格", "人格"},
			},
			Synonyms: map[string]string{
				"identity": "persona", "profile": "persona",
				"身份": "人设", "设定": "人设",
			},
			Examples: []string{"persona set friendly assistant", "人设 清除"},
		},
		{
			Name:        "forget",
			Description: "Clear conversation history or memory",
			Intent:      IntentForget,
			Priority:    60,
			Enabled:     true,
			Builtin:     true,
			Keywords: map[string][]string{
				"en": {"forget", "clear", "reset", "wipe"},
				"zh": {"忘记", "清除", "重置", "清空"},
			},
			Synonyms: map[string]string{
				"erase": "clear", "flush": "clear",
				"清除记忆": "清除", "忘掉": "忘记", "删掉": "清除",
			},
			Examples: []string{"forget", "清除记忆"},
		},
		{
			Name:        "query-status",
			Description: "Query the status of a process or system",
			Intent:      IntentQueryStatus,
			Priority:    70,
			Enabled:     true,
			Builtin:     true,
			Keywords: map[string][]string{
				"en": {"status", "state", "health", "check", "info"},
				"zh": {"状态", "情况", "健康", "检查", "信息"},
			},
			Synonyms: map[string]string{
				"stat": "status", "how": "status",
				"怎么样": "状态", "运行状况": "状态", "跑得怎样": "状态",
			},
			Examples: []string{"status worker1", "查看所有进程状态"},
		},
		{
			Name:        "query-list",
			Description: "List processes, tasks, or resources",
			Intent:      IntentQueryList,
			Priority:    80,
			Enabled:     true,
			Builtin:     true,
			Keywords: map[string][]string{
				"en": {"list", "show", "ls", "all", "enumerate"},
				"zh": {"列表", "显示", "列出", "查看", "所有"},
			},
			Synonyms: map[string]string{
				"dir": "list", "display": "show",
				"展示": "显示", "罗列": "列出",
			},
			Examples: []string{"list all workers", "显示所有进程"},
		},
	}
}
