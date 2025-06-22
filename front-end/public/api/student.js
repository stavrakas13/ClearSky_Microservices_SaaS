// student.js
import { request } from './_request.js';

/**
 * Submit a new review request.
 * orchestrator: PATCH /student/reviewRequest
 */
export const postReviewRequest = ({ course_id, exam_period, student_message }) =>
  request('/student/reviewRequest', {
    method : 'PATCH',
    body   : { course_id, exam_period, student_message }
  });

/**
 * Check review status.
 * orchestrator: PATCH /student/status
 */
export const getReviewStatus = ({ course_id, exam_period }) =>
  request('/student/status', {
    method : 'PATCH',
    body   : { course_id, exam_period }
  });
