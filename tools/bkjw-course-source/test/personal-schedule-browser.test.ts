import { expect, it } from "vitest"
import { chromium } from "playwright"
import { openPersonalSchedule, readPersonalCourses } from "../src/personal-schedule.js"

// Opt in when Chromium is installed (also bundled in the project Docker image).
// All upstream requests are intercepted; no school account or cookies are used.
it.runIf(process.env.BKJW_BROWSER_TESTS === "1")("loads a collapsed EAMS menu, waits for its calendar, and reads each selected semester", async () => {
  const browser = await chromium.launch({headless: true})
  try {
    const context = await browser.newContext()
    const queries: string[] = []
    await context.route("http://bkjw.chd.edu.cn/**", async route => {
      if (new URL(route.request().url()).pathname.endsWith("!courseTable.action")) {
        const params = new URLSearchParams(route.request().postData() ?? "")
        expect(params.get("setting.kind")).toBe("std")
        const id = params.get("semester.id") ?? ""
        queries.push(id)
        await route.fulfill({json: {activities: id === "101" ? [
          {courseName: "程序设计", lessonNo: "CS1.01", courseId: "123", teacherName: "甲,乙"},
          {courseName: "程序设计", lessonNo: "CS1.01", courseId: "123", teacherName: "乙,甲"},
        ] : []}})
        return
      }
      await route.fulfill({contentType: "text/html", body: `<!doctype html><html><body>
        <nav style="display:none"><a href="/eams/courseTableForStd.action" id="menu">我的课表</a></nav>
        <div id="contentDiv"></div>
        <script>
          document.querySelector('#menu').onclick = function(e) {
            e.preventDefault(); setTimeout(function() {
              document.querySelector('#contentDiv').innerHTML = '<select id="courseTableType"><option value="std">个人</option><option value="class">班级</option></select><input id="semesterCalendar_target" type="hidden" value="101"><input class="calendar-text" id="querySemester" readonly value="2026-2027学年第1学期"><div id="calendar"></div>';
              window.searchTable = function() {
                var params = new URLSearchParams({'setting.kind':document.querySelector('#courseTableType').value,'semester.id':document.querySelector('#semesterCalendar_target').value});
                fetch('/eams/courseTableForStd!courseTable.action',{method:'POST',body:params}).then(r=>r.json()).then(data=>{window.table0=data;});
              };
              document.querySelector('#querySemester').onclick = function() {
                setTimeout(function() {
                  document.querySelector('#calendar').innerHTML = '<table id="semesterCalendar_yearTb"><tr><td>2026-2027</td></tr></table><table id="semesterCalendar_termTb"><tr><td val="101">1</td><td val="102">2</td></tr></table>';
                },300);
              };
            },100);
          };
        </script></body></html>`})
    })
    const page = await context.newPage()
    const semesters = await openPersonalSchedule(page)
    expect(semesters.map(s => s.normalized).sort()).toEqual(["2026-2027-1", "2026-2027-2"])
    const first = semesters.find(s => s.id === "101")!
    const second = semesters.find(s => s.id === "102")!
    const signal = new AbortController().signal
    expect(await readPersonalCourses(page, first, signal)).toEqual([{lesson_no:"CS1.01", source_lesson_id:"123", name:"程序设计"}])
    // An empty next semester must not reuse table0 from the previous response.
    expect(await readPersonalCourses(page, second, signal)).toEqual([])
    expect(queries).toEqual(["101", "102"])
  } finally { await browser.close() }
}, 30000)
