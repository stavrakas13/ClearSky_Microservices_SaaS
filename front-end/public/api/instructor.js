// front-end/public/api/instructor.js
import { request } from './_request.js';

/**
 * Fetch the list of pending review requests for the instructor.
 * Endpoint: PATCH /instructor/review-list
 *
 * @param {Object} payload
 * @param {string=} payload.course_id
 * @param {string=} payload.exam_period
 * @returns {Promise<Array>}  Array of review objects
 */
export const getPendingReviews = async ({ course_id, exam_period } = {}) => {
  const res = await request('/instructor/review-list', {
    method: 'PATCH',
    body: { course_id, exam_period }
  });

  /* The orchestrator wraps the service response:
     { data: { message, data: [...] } }
     └─ we want the inner `data` array.                             */
  return res?.data?.data ?? res?.data ?? res;
};

/**
 * Send an instructor’s reply to a pending request.
 * Endpoint: PATCH /instructor/reply
 *
 * @param {Object} payload
 * @param {string}  payload.user_id
 * @param {string}  payload.course_id
 * @param {string}  payload.exam_period
 * @param {string}  payload.instructor_reply_message
 * @param {string}  payload.instructor_action
 * @returns {Promise<Object>}  API response
 */
export const postInstructorReply = ({
  user_id,
  course_id,
  exam_period,
  instructor_reply_message,
  instructor_action
}) =>
  request('/instructor/reply', {
    method: 'PATCH',
    body: {
      user_id,
      course_id,
      exam_period,
      instructor_reply_message,
      instructor_action
    }
  });
