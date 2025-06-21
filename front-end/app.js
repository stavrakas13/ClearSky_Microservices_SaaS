// file: front-end/app.js
import express           from 'express';
import session           from 'express-session';
import path              from 'path';
import { fileURLToPath } from 'url';
import morgan            from 'morgan';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const app       = express();

// 1) Request logging
app.use(morgan('dev'));

// 2) Serve static files from /public
app.use(express.static(path.join(__dirname, 'public')));

// 3) Sessions & template locals
app.use(session({
  secret           : 'change-this-secret',
  resave           : false,
  saveUninitialized: true,
}));
app.use((req, res, next) => {
  res.locals.user       = req.session.user || null;
  res.locals.currentUrl = req.originalUrl;
  next();
});

// 4) EJS views setup
app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'views'));

// 5) Auth helper
function auth(role) {
  return (req, res, next) => {
    if (!req.session.user) return res.redirect('/login');
    if (role && req.session.user.role !== role)
      return res.redirect(`/${req.session.user.role}`);
    next();
  };
}

// 6) Dummy local users (dev fallback)
const users = { alice: 'student', bob: 'instructor', iris: 'institution' };

// 7) UI routes

// Home
app.get('/', (req, res) => {
  if (!req.session.user) return res.redirect('/login');
  res.redirect(`/${req.session.user.role}`);
});

// Signup / Login
app.get('/signup', (_, res) =>
  res.render('institution/userManagement', { user: null, title: 'Sign Up' })
);
app.get('/login', (_, res) =>
  res.render('login', { title: 'Log in', error: null, user: null })
);
app.post('/login', (req, res) => {
  const { username, password } = req.body;
  const role = users[username];
  if (!role || password !== '1234') {
    return res.render('login', { title: 'Log in', error: 'Invalid credentials', user: null });
  }
  req.session.user = { username, role };
  res.redirect(`/${role}`);
});
app.get('/logout', (req, res) =>
  req.session.destroy(() => res.redirect('/login'))
);

// Student UI
app.get('/student',            auth('student'), (req, res) => res.render('student/dashboard',     { user: req.session.user, title: 'Dashboard' }));
app.get('/student/statistics', auth('student'), (req, res) => res.render('student/statistics',    { user: req.session.user, title: 'Statistics' }));
app.get('/student/my-courses', auth('student'), (req, res) => res.render('student/myCourses',     { user: req.session.user, title: 'My Courses' }));
app.get('/student/request',    auth('student'), (req, res) => res.render('student/reviewRequest', { user: req.session.user, title: 'Review Request' }));
app.get('/student/status',     auth('student'), (req, res) => res.render('student/reviewStatus',  { user: req.session.user, title: 'Review Status' }));
app.get('/student/personal',   auth('student'), (req, res) => res.render('student/personal',      { user: req.session.user, title: 'Personal Grades' }));

// Instructor UI
app.get('/instructor',              auth('instructor'), (req, res) => res.render('instructor/dashboard',   { user: req.session.user, title: 'Dashboard' }));
app.get('/instructor/post-initial', auth('instructor'), (req, res) => res.render('instructor/postInitial', { user: req.session.user, title: 'Post Initial' }));
app.get('/instructor/post-final',   auth('instructor'), (req, res) => res.render('instructor/postFinal',   { user: req.session.user, title: 'Post Final' }));
app.get('/instructor/review-list',  auth('instructor'), (req, res) => res.render('instructor/reviewList',  { user: req.session.user, title: 'Review Requests' }));
app.get('/instructor/reply',        auth('instructor'), (req, res) => {
  const request_id = req.query.req || '';
  res.render('instructor/replyForm', {
    user         : req.session.user,
    title        : 'Reply to Review Request',
    request_id,
    course_name  : 'software II',
    exam_period  : 'spring 2025',
    student_name : 'john doe',
  });
});
app.get('/instructor/statistics',   auth('instructor'), (req, res) => res.render('instructor/statistics',  { user: req.session.user, title: 'Statistics' }));

// Institution UI
app.get('/institution',                 auth('institution'), (req, res) => res.render('institution/dashboard',       { user: req.session.user, title: 'Dashboard' }));
app.get('/institution/register',        auth('institution'), (req, res) => res.render('institution/register',        { user: req.session.user, title: 'Register' }));
app.get('/institution/purchase',        auth('institution'), (req, res) => res.render('institution/purchase',        { user: req.session.user, title: 'Purchase' }));
app.get('/institution/user-management', auth('institution'), (req, res) => res.render('institution/userManagement', { user: req.session.user, title: 'Users' }));
app.get('/institution/statistics',      auth('institution'), (req, res) => res.render('institution/statistics',     { user: req.session.user, title: 'Statistics' }));

// 8) Fallback body parsing for form submissions
app.use(express.urlencoded({ extended: false }));
app.use(express.json());

// 9) Start server
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => console.log(`✔ Front-end listening on http://localhost:${PORT}`));
