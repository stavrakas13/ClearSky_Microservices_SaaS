// instructor.js
import { request } from './_request.js';

export const getPendingReviews = ({ course_id, exam_period }) =>
  request('/instructor/review-list', {
    method : 'PATCH',
    body   : { course_id, exam_period }
  });

export const postInstructorReply = ({
  user_id, course_id, exam_period, instructor_reply_message, instructor_action
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
