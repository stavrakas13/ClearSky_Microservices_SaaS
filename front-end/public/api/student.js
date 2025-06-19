// student.js
import { request } from './_request.js';

export const postReviewRequest = ({ user_id, course_id, exam_period, student_message }) =>
  request('/student/reviewRequest', {
    method : 'PATCH',
    body   : { user_id, course_id, exam_period, student_message }
  });

export const getReviewStatus = ({ user_id, course_id, exam_period }) =>
  request('/student/status', {
    method : 'PATCH',
    body   : { user_id, course_id, exam_period }
  });
