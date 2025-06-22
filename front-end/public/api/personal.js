// personal.js
import { request } from './_request.js';

/**
 * Fetch all courses & periods for the logged-in student.
 * (orchestrator: GET /stats/available → returns all submission logs)
 */
export const getStudentCourses = () =>
  request('/stats/available');

/**
 * Fetch personal grades for a given course & exam period.
 * (orchestrator: GET /personal/grades?course_id=…&exam_period=…)
 */
export const getPersonalGrades = ({ course_id, exam_period }) =>
  request(`/personal/grades?course_id=${encodeURIComponent(course_id)}&exam_period=${encodeURIComponent(exam_period)}`);
