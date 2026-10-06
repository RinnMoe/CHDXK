import { expect, it } from "vitest"
import { chromium } from "playwright"
import { openPersonalSchedule, readPersonalCourses } from "../src/personal-schedule.js"

// Opt in when Chromium is installed (also bundled in the project Docker image).
// All upstream requests are intercepted; no school account or cookies are used.
it.runIf(process.env.BKJW_BROWSER_TESTS === "1").each([
  {storedCalendar:false, failure:0, permanent:false, cancel:false},
  {storedCalendar:true, failure:503, permanent:false, cancel:false},
  {storedCalendar:true, failure:401, permanent:false, cancel:false},
  {storedCalendar:true, failure:503, permanent:true, cancel:false},
  {storedCalendar:true, failure:503, permanent:false, cancel:true},
])("reads semesters without stale data and bounds recovery: %j", async ({storedCalendar,failure,permanent,cancel}) => {
  const browser = await chromium.launch({headless: true})
  try {
    const context = await browser.newContext()
    const queries: string[] = []
    let secondRequests = 0
    await context.route("http://bkjw.chd.edu.cn/**", async route => {
      if (new URL(route.request().url()).pathname.endsWith("!courseTable.action")) {
        const params = new URLSearchParams(route.request().postData() ?? "")
        expect(params.get("setting.kind")).toBe("std")
        const id = params.get("semester.id") ?? ""
        queries.push(id)
        if (id === "102" && ++secondRequests && failure && (permanent || secondRequests === 1)) {
          await route.fulfill({status:failure, contentType:"text/plain", body:"upstream unavailable"})
          return
        }
        await route.fulfill({json: {activities: id === "101" ? [[
          {courseName: "程序设计", lessonNo: "CS1.01", courseId: "123", teacherName: "甲,乙"},
          {courseName: "程序设计", lessonNo: "CS1.01", courseId: "123", teacherName: "乙,甲"},
        ], []] : [[], []]}})
        return
      }
      await route.fulfill({contentType: "text/html", body: `<!doctype html><html><body>
        <nav style="display:none"><a href="/eams/courseTableForStd.action" id="menu">我的课表</a></nav>
        <div id="contentDiv"></div>
        <script>
          ${storedCalendar ? `window.jQuery = function() { return {data:function() { return window.semesters }} };` : ""}
          document.querySelector('#menu').onclick = function(e) {
            e.preventDefault(); setTimeout(function() {
              document.querySelector('#contentDiv').innerHTML = '<select id="courseTableType"><option value="std">个人</option><option value="class">班级</option></select><input id="semesterCalendar_target" type="hidden" value="101"><input class="calendar-text" id="querySemester" readonly value="2026-2027学年第1学期"><div id="calendar"></div>';
              window.searchTable = function() {
                var params = new URLSearchParams({'setting.kind':document.querySelector('#courseTableType').value,'semester.id':document.querySelector('#semesterCalendar_target').value});
                fetch('/eams/courseTableForStd!courseTable.action',{method:'POST',body:params}).then(r=>r.ok?r.json():null).then(data=>{if(data) window.table0=data;});
              };
              ${storedCalendar ? `setTimeout(function() {window.semesters={y0:[{id:101,schoolYear:'2026-2027',name:'1'},{id:102,schoolYear:'2026-2027',name:'2'}]};},300);` : ""}
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
    const controller = new AbortController()
    const signal = controller.signal
    expect(await readPersonalCourses(page, first, signal)).toEqual([{lesson_no:"CS1.01", source_lesson_id:"123", name:"程序设计"}])
    // An empty next semester must not reuse table0 from the previous response.
    const retries: number[] = []
    const reading = readPersonalCourses(page, second, signal, attempt => {retries.push(attempt); if(cancel) controller.abort()})
    if (cancel) await expect(reading).rejects.toThrow("CANCELLED")
    else if (failure===401) await expect(reading).rejects.toThrow("AUTH_REQUIRED")
    else if (permanent) await expect(reading).rejects.toThrow("PERSONAL_REQUEST_FAILED")
    else expect(await reading).toEqual([])
    expect(retries).toEqual(cancel ? [1] : permanent ? [1,2] : failure===503 ? [1] : [])
    expect(queries).toEqual(permanent ? ["101","102","102","102"] : failure===503 && !cancel ? ["101","102","102"] : ["101","102"])
  } finally { await browser.close() }
}, 30000)
