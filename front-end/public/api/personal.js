// front-end/public/api/personal.js
import { request } from './_request.js';

/**
 * Fetch all courses & periods for the logged-in student.
 * (orchestrator: GET /stats/available → returns every submission log)
 */
export const getStudentCourses = () =>
  request('/stats/available');

/**
 * Fetch personal grades for a given course and (optionally) exam period.
 * If exam_period is omitted the backend will return the most recent one.
 *
 * @param {{ course_id: string|number, exam_period?: string }} params
 */
export const getPersonalGrades = ({ course_id, exam_period }) => {
  const qs = new URLSearchParams();
  if (course_id !== undefined && course_id !== null) qs.append('course_id', course_id);
  if (exam_period !== undefined && exam_period !== null) qs.append('exam_period', exam_period);

  // empty query-string → “/personal/grades”
  const suffix = qs.toString() ? `?${qs}` : '';
  return request(`/personal/grades${suffix}`);
};
