// front-end/public/js/student/my-courses.js
//
// Shows all courses for the logged-in student and builds action links
// that carry ONLY the numeric course_id (e.g. 3205).
// ────────────────────────────────────────────────────────────────────
import { flash } from '../../script.js';
import { getStudentCourses } from '../../api/personal.js';

window.addEventListener('DOMContentLoaded', async () => {
  try {
    const courses = await getStudentCourses();         // array from /personal/grades
    const tbody   = document.querySelector('table tbody');

    if (!courses.length) {
      tbody.innerHTML = `
        <tr><td colspan="4" style="text-align:center;">No courses found.</td></tr>`;
      return;
    }

    tbody.innerHTML = courses.map(c => {
      /* ── 1) Resolve display fields ─────────────────────────────── */
      const courseName = c.classTitle        ?? c.course_name  ?? '—';
      const examPeriod = c.declarationPeriod ?? c.exam_period  ?? '—';

      let status = c.status ?? c.grading_status;
      if (typeof status === 'number') status = status === 0 ? 'open' : 'closed';
      if (!status) status = '—';

      /* ── 2) Derive a clean numeric course-id for links ─────────── */
      const rawId = c.course_id ?? c.id ?? courseName;  // whatever we’ve got
      const numericId =
        String(rawId).match(/\((\d+)\)/)?.[1]   // digits inside ( ... )
        || String(rawId).replace(/\D/g, '')     // else keep any digits
        || encodeURIComponent(rawId);           // fallback

      /* ── 3) Build row ──────────────────────────────────────────── */
      return `
        <tr ${status === 'open' ? 'style="background:#e6e7ea;"' : ''}>
          <td>${courseName}</td>
          <td>${examPeriod}</td>
          <td>${status}</td>
          <td>
            <a class="button"
               href="/student/personal?course=${numericId}&period=${encodeURIComponent(examPeriod)}">
              View grades
            </a>

            <a class="button${status !== 'open' ? ' button--secondary' : ''}"
               href="/student/request?course=${numericId}&period=${encodeURIComponent(examPeriod)}">
              Ask review
            </a>

            <a class="button${status === 'open' ? ' button--secondary' : ''}"
               href="/student/status?course=${numericId}&period=${encodeURIComponent(examPeriod)}">
              Status
            </a>
          </td>
        </tr>`;
    }).join('');
  } catch (err) {
    flash(err.message || 'Failed to load courses');
  }
});
