// personal.js
import { request } from './_request.js';

export const getStudentCourses = ({ user_id }) =>
  request('/personal/courses', { method: 'POST', body: { user_id } });

export const getPersonalGrades = ({ user_id, course_id, exam_period }) =>
  request('/personal/grades', {
    method : 'POST',
    body   : { user_id, course_id, exam_period }
  });
