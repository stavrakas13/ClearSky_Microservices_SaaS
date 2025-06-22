// public/api/instructor.js
import { request } from './_request.js';

/**
 * Fetch pending review requests for the current instructor.
 * orchestrator: PATCH /instructor/review-list
 */
export const getPendingReviews = ({ course_id, exam_period }) =>
  request('/instructor/review-list', {
    method : 'PATCH',
    body   : { course_id, exam_period }
  });

/**
 * Send an instructor’s reply to a specific student request.
 * orchestrator: PATCH /instructor/reply
 */
export const postInstructorReply = ({
  user_id,                // student’s user_id from URL
  course_id,
  exam_period,
  instructor_reply_message,
  instructor_action
}) =>
  request('/instructor/reply', {
    method : 'PATCH',
    body   : {
      user_id,
      course_id,
      exam_period,
      instructor_reply_message,
      instructor_action
    }
  });
