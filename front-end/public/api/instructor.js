// front-end/public/api/instructor.js
import { request } from './_request.js';

/*-------------------------------------------------------------*
 | Utilities                                                   |
 *-------------------------------------------------------------*/

/**
 * Strip undefined, null, or empty‐string values from an object,
 * returning a *new* object that only contains “meaningful” keys.
 */
const prune = (obj = {}) =>
  Object.fromEntries(
    Object.entries(obj).filter(
      ([, v]) => v !== undefined && v !== null && v !== ''
    )
  );

/*-------------------------------------------------------------*
 | API calls                                                   |
 *-------------------------------------------------------------*/

/**
 * Get pending-review requests for an instructor.
 * PATCH /instructor/review-list
 *
 * @param {{course_id?:string, exam_period?:string}=} filters
 * @returns {Promise<Array>}
 */
export const getPendingReviews = async (filters = {}) => {

  // Log filters before pruning and sending
  console.log('[DEBUG] 🟡 getPendingReviews called with filters:', filters);
  const prunedBody = prune(filters);

  // Log pruned body actually sent in request
  console.log('[DEBUG] 🟡 Pruned request body:', prunedBody);


  const res = await request('/instructor/review-list', {
    method: 'PATCH',
    body: prune(filters)           // <-- ⬅⬅⬅  IMPORTANT LINE
  });
  
  console.log('[DEBUG] 🟡 Response from /instructor/review-list:', res);

  return res?.data?.data ?? res?.data ?? res;
};

/**
 * Send an instructor reply.
 * PATCH /instructor/reply
 *
 * @param {{
 *   user_id: string,
 *   course_id: string,
 *   exam_period?: string,
 *   instructor_reply_message: string,
 *   instructor_action: string
 * }} payload
 */
export const postInstructorReply = (payload) =>
  request('/instructor/reply', {
    method: 'PATCH',
    body: prune(payload)           // <-- ⬅⬅⬅  IMPORTANT LINE
  });
