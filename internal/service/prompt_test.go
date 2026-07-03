package service

import (
	"regexp"
	"strings"
	"testing"
)

func TestBuildSystemMessage_WithEmotion(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildSystemMessage(true, false)

	if !strings.Contains(msg, "情绪标注") {
		t.Error("expected emotion section in system message")
	}
	if !strings.Contains(msg, "@提及识别") {
		t.Error("expected mention section in system message")
	}
	if !strings.Contains(msg, "语气保真与情绪标注") {
		t.Error("expected rule5 title with emotion")
	}
	for _, tag := range []string{"[有品位]", "[崇尚行动]", "[使命必达]", "[尚方宝剑]"} {
		if !strings.Contains(msg, tag) {
			t.Errorf("expected custom emoji tag %s in system message", tag)
		}
	}
}

func TestBuildSystemMessage_NoEmotion(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildSystemMessage(false, false)

	if strings.Contains(msg, "情绪标注(⚠️ 必须执行)") {
		t.Error("expected no emotion section")
	}
	if !strings.Contains(msg, "@提及识别") {
		t.Error("expected mention section")
	}
	if strings.Contains(msg, "语气保真与情绪标注") {
		t.Error("expected rule5 title without emotion keyword")
	}
	if !strings.Contains(msg, "语气保真") {
		t.Error("expected rule5 title")
	}
}

func TestBuildSystemMessage_SkipMention(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildSystemMessage(true, true)

	if strings.Contains(msg, "@提及识别") {
		t.Error("expected no mention section when skipMention=true")
	}
}

func TestBuildSystemMessage_BroadcastMention(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildSystemMessage(true, false)

	if !strings.Contains(msg, "### 广播 @mention") {
		t.Error("expected broadcast mention section in system message")
	}
	if !strings.Contains(msg, "@所有人") {
		t.Error("expected @所有人 broadcast token guidance")
	}
	if !strings.Contains(msg, "@所有AI") {
		t.Error("expected @所有AI broadcast token guidance")
	}
	if strings.Contains(msg, "@所有 AI") {
		t.Error("broadcast token must be atomic: @所有AI must never contain a space (@所有 AI)")
	}
}

func TestBuildSystemMessage_SkipBroadcastMention(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildSystemMessage(true, true)

	if strings.Contains(msg, "### 广播 @mention") {
		t.Error("expected no broadcast mention section when skipMention=true")
	}
}

func TestBuildSystemMessage_EditOnlyMention(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildSystemMessage(true, false, "edit_only")

	if !strings.Contains(msg, "@提及识别") {
		t.Error("expected mention section in edit_only system message")
	}
	if !strings.Contains(msg, "### 广播 @mention") {
		t.Error("expected broadcast mention section in edit_only system message")
	}
}

func TestBuildSystemMessage_EditOnlySkipMention(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildSystemMessage(true, true, "edit_only")

	if strings.Contains(msg, "@提及识别") {
		t.Error("expected no mention section in edit_only when skipMention=true")
	}
}

func TestBuildSystemMessage_EditOnlyWithEmotion(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildSystemMessage(true, false, "edit_only")

	if !strings.Contains(msg, "### 规则 5:情绪标注") {
		t.Error("expected emotion annotation rule (rule 5) in edit_only system message when emotion enabled")
	}
	if !strings.Contains(msg, "情绪标注(⚠️ 必须执行)") {
		t.Error("expected emotion annotation section body in edit_only system message when emotion enabled")
	}
	for _, tag := range []string{"[有品位]", "[崇尚行动]", "[使命必达]", "[尚方宝剑]"} {
		if !strings.Contains(msg, tag) {
			t.Errorf("expected custom emoji tag %s in edit_only system message", tag)
		}
	}
}

func TestBuildSystemMessage_EditOnlyNoEmotion(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildSystemMessage(false, false, "edit_only")

	// When emotion is disabled, the entire rule 5 (heading + intro + section)
	// must be gone, not just the placeholder body. Guards against regression
	// where the rule 5 heading/intro was hardcoded in systemPromptEditOnly.
	if strings.Contains(msg, "### 规则 5:情绪标注") {
		t.Error("expected no emotion annotation rule (rule 5) in edit_only when emotion disabled")
	}
	if strings.Contains(msg, "情绪标注(⚠️ 必须执行)") {
		t.Error("expected no emotion annotation section in edit_only when emotion disabled")
	}
}

func TestBuildSystemMessage_FallbackToCustomSystem(t *testing.T) {
	ResetPromptsToDefaults()
	// Restore global prompt state afterwards so a shuffled run order cannot
	// leak this custom override into other tests that rely on the default
	// template (e.g. TestTranscribe_EditWithBuffer_HasEditorSection).
	defer ResetPromptsToDefaults()
	activePrompts.System = "custom system prompt"
	activePrompts.SystemOverridden = true
	// AppendOnlyOverridden and EditOnlyOverridden remain false

	appendMsg := BuildSystemMessage(true, false, "append")
	if appendMsg != "custom system prompt" {
		t.Errorf("append mode should fall back to custom system prompt, got: %s", appendMsg[:50])
	}

	editOnlyMsg := BuildSystemMessage(true, false, "edit_only")
	if editOnlyMsg != "custom system prompt" {
		t.Errorf("edit_only mode should fall back to custom system prompt, got: %s", editOnlyMsg[:50])
	}
}

func TestBuildUserMessage_AppendNoContext(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildUserMessage("append", "", "", true)

	if !strings.Contains(msg, "请转写音频中的语音") {
		t.Error("expected transcribe task")
	}
}

func TestBuildUserMessage_AppendWithContext(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildUserMessage("append", "existing text", "", true)

	if !strings.Contains(msg, "<input_buffer>") {
		t.Error("expected input_buffer tag")
	}
	if !strings.Contains(msg, "existing text") {
		t.Error("expected context text")
	}
}

func TestBuildUserMessage_AppendWithVocab(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildUserMessage("append", "existing", "vocab data", true)

	if !strings.Contains(msg, "<vocabulary_reference>") {
		t.Error("expected vocabulary_reference tag")
	}
	if !strings.Contains(msg, "vocab data") {
		t.Error("expected vocabulary content")
	}
}

func TestBuildUserMessage_EditWithContext(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildUserMessage("edit", "text to edit", "", true)

	if !strings.Contains(msg, "当前已有的文本") {
		t.Error("expected edit input buffer template")
	}
	if !strings.Contains(msg, "text to edit") {
		t.Error("expected context text")
	}
	if !strings.Contains(msg, "编辑指令") {
		t.Error("expected edit task instruction")
	}
}

func TestBuildUserMessage_EditOnly(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildUserMessage("edit_only", "text to edit", "", true)

	if !strings.Contains(msg, "根据音频语音指令编辑上方 input_buffer") {
		t.Error("expected edit_only task instruction")
	}
}

func TestBuildUserMessage_EditNoEmotion(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildUserMessage("edit", "text", "", false)

	if strings.Contains(msg, "情绪标注") {
		t.Error("expected no emotion annotation in no-emotion mode")
	}
	if !strings.Contains(msg, "语气保真") {
		t.Error("expected tone preservation")
	}
}

func TestBuildUserMessage_DefaultMode(t *testing.T) {
	ResetPromptsToDefaults()
	msg := BuildUserMessage("", "", "", true)

	if !strings.Contains(msg, "请转写音频中的语音") {
		t.Error("expected transcribe task for default mode")
	}
}

func TestBuildVocabularyReference_NoPersonalNoMember(t *testing.T) {
	result := BuildVocabularyReference("", "", "chat data")
	if result != "chat data" {
		t.Errorf("expected plain chat context, got %q", result)
	}
}

func TestBuildVocabularyReference_WithPersonal(t *testing.T) {
	result := BuildVocabularyReference("personal ctx", "", "chat data")
	if !strings.Contains(result, "<personal_vocabulary>") {
		t.Error("expected personal_vocabulary tag")
	}
	if !strings.Contains(result, "<latest_chat_context>") {
		t.Error("expected latest_chat_context tag")
	}
}

func TestBuildVocabularyReference_WithMember(t *testing.T) {
	result := BuildVocabularyReference("", "member ctx", "")
	if !strings.Contains(result, "<member_vocabulary>") {
		t.Error("expected member_vocabulary tag")
	}
}

func TestBuildVocabularyReference_All(t *testing.T) {
	result := BuildVocabularyReference("personal", "member", "chat")
	if !strings.Contains(result, "<personal_vocabulary>") {
		t.Error("expected personal_vocabulary")
	}
	if !strings.Contains(result, "<member_vocabulary>") {
		t.Error("expected member_vocabulary")
	}
	if !strings.Contains(result, "<latest_chat_context>") {
		t.Error("expected latest_chat_context")
	}
}

func TestBuildSystemMessage_ASRCleanup_AppendAndTemplate(t *testing.T) {
	ResetPromptsToDefaults()
	// v2 方案改动 A/B/D 必须同时落到 append 与 template(edit)两模板;
	// 用 emotion 开/关两种渲染都验证一遍,防止 rule5 段被 emotion 分支吃掉。
	for _, emotion := range []bool{true, false} {
		for _, mode := range []string{"append", "edit"} {
			var msg string
			if mode == "append" {
				msg = BuildSystemMessage(emotion, false)
			} else {
				msg = BuildSystemMessage(emotion, false, "edit")
			}
			if !strings.Contains(msg, "语义冗余的实义词重复可去重合并") {
				t.Errorf("[emotion=%v mode=%s] missing 改动A 例外条款", emotion, mode)
			}
			if !strings.Contains(msg, "合并三前提") {
				t.Errorf("[emotion=%v mode=%s] missing 改动B 合并三前提", emotion, mode)
			}
			if !strings.Contains(msg, "同意群内最小语序整理") {
				t.Errorf("[emotion=%v mode=%s] missing 改动B 最小语序整理", emotion, mode)
			}
			if !strings.Contains(msg, "在规则4 授权范围内") {
				t.Errorf("[emotion=%v mode=%s] missing 改动D 最高层新口径", emotion, mode)
			}
			if strings.Contains(msg, "润色只限于去除冗余,不得改变用词、句序、术语") {
				t.Errorf("[emotion=%v mode=%s] 旧最高层口径未被替换,会压回规则4", emotion, mode)
			}
			if !strings.Contains(msg, "示例19 - ASR 语义冗余去重合并") {
				t.Errorf("[emotion=%v mode=%s] missing 示例19", emotion, mode)
			}
			if !strings.Contains(msg, "@Thomas.fu 创建两个子区分别跟踪解决这两个 issue") {
				t.Errorf("[emotion=%v mode=%s] 示例19 正例文本缺失", emotion, mode)
			}
			// 放宽后必须出现"同一分配语义"授权口径 + 分配语义关键词"分别"/"都"/"也"
			if !strings.Contains(msg, "同一分配语义") {
				t.Errorf("[emotion=%v mode=%s] 规则4 缺放宽后的\"同一分配语义\"授权口径", emotion, mode)
			}
			if strings.Contains(msg, "同一动作(同一动词/近义动词族)") {
				t.Errorf("[emotion=%v mode=%s] 旧\"同一动作/近义动词族\"口径未替换", emotion, mode)
			}
			// OCT-102:示例19 判据段必须与放宽后规则4 正文对齐——不得留旧"同一动作族"口径,且要显式出现"同一分配语义"。
			const ex19Head = "示例19 - ASR 语义冗余去重合并"
			const ex19Tail = "示例19 反例"
			if i, j := strings.Index(msg, ex19Head), strings.Index(msg, ex19Tail); i >= 0 && j > i {
				rationale := msg[i:j]
				if strings.Contains(rationale, "同一动作族") {
					t.Errorf("[emotion=%v mode=%s] 示例19 判据仍含旧\"同一动作族\"口径,与放宽后规则4 自相矛盾", emotion, mode)
				}
				if !strings.Contains(rationale, "同一分配语义") {
					t.Errorf("[emotion=%v mode=%s] 示例19 判据缺\"同一分配语义\"字样,与规则正文口径未对齐", emotion, mode)
				}
			} else {
				t.Errorf("[emotion=%v mode=%s] 未定位到示例19 判据段(head=%d tail=%d)", emotion, mode, i, j)
			}
			// nit1 回归:示例19 ✅ 行不得含反斜杠双引号(raw string 里 \\\" 会原样进 prompt)
			if strings.Contains(msg, `他刚才说\"口令就是`) {
				t.Errorf("[emotion=%v mode=%s] 示例19 ✅ 行残留反斜杠双引号(nit1 回归)", emotion, mode)
			}
		}
	}
}

func TestBuildSystemMessage_ASRCleanup_EditOnly(t *testing.T) {
	ResetPromptsToDefaults()
	// 改动 E:editOnly 规则3 授权 + 反例
	for _, emotion := range []bool{true, false} {
		msg := BuildSystemMessage(emotion, false, "edit_only")
		if !strings.Contains(msg, "同一分配语义") {
			t.Errorf("[emotion=%v] editOnly 规则3 缺 分配语义授权口径", emotion)
		}
		if strings.Contains(msg, "与转写模式规则4/规则5 例外对齐") {
			t.Errorf("[emotion=%v] editOnly 规则3 仍在引用不可见的转写规则,应自洽表述", emotion)
		}
		if !strings.Contains(msg, "引用原话不合并") {
			t.Errorf("[emotion=%v] editOnly 规则3 缺 改动E 反例", emotion)
		}
		if !strings.Contains(msg, "施事不同不合并") {
			t.Errorf("[emotion=%v] editOnly 规则3 缺 改动E 反例(施事)", emotion)
		}
	}
}

func TestBuildSystemMessage_ExampleNumbersUnique(t *testing.T) {
	ResetPromptsToDefaults()
	// 防止改动C 新增示例19 与 emotion 示例13-18 撞车,或与语言润色 1-12 重叠。
	// 只统计"示例N -"(严格带破折号)的出现,避开正文里"示例19"字样的行内引用。
	re := regexp.MustCompile(`示例(\d+)\s*-`)
	for _, emotion := range []bool{true, false} {
		for _, mode := range []string{"append", "edit", "edit_only"} {
			var msg string
			if mode == "append" || mode == "edit" {
				msg = BuildSystemMessage(emotion, false, mode)
			} else {
				msg = BuildSystemMessage(emotion, false, "edit_only")
			}
			seen := map[string]int{}
			for _, m := range re.FindAllStringSubmatch(msg, -1) {
				seen[m[1]]++
			}
			for n, c := range seen {
				if c > 1 {
					t.Errorf("[emotion=%v mode=%s] 示例%s- 出现 %d 次,应唯一", emotion, mode, n, c)
				}
			}
			if emotion && (mode == "append" || mode == "edit") {
				// emotion 开 + append/edit:必须同时有 12 和 19,且 13-18 都在
				for _, want := range []string{"12", "13", "14", "15", "16", "17", "18", "19"} {
					if seen[want] == 0 {
						t.Errorf("[emotion=true mode=%s] 缺示例%s", mode, want)
					}
				}
			}
			if !emotion && (mode == "append" || mode == "edit") {
				// emotion 关:示例19 仍然在(它属于语言润色示例区,不属于情绪示例)
				if seen["19"] == 0 {
					t.Errorf("[emotion=false mode=%s] 缺示例19 (改动C 应在 emotion 关时也渲染)", mode)
				}
				if seen["13"] != 0 || seen["18"] != 0 {
					t.Errorf("[emotion=false mode=%s] 情绪示例13-18 不应出现", mode)
				}
			}
		}
	}
}

func TestIsNoSpeech(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"", true},
		{"[NO_SPEECH]", true},
		{"  [NO_SPEECH]  ", true},
		{"some text [NO_SPEECH] more", true},
		{"hello world", false},
		{"NOSPEECH", false},
	}

	for _, tt := range tests {
		got := IsNoSpeech(tt.text)
		if got != tt.want {
			t.Errorf("IsNoSpeech(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
}
