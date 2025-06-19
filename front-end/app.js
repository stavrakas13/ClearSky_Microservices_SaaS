import express                     from 'express';
import session                     from 'express-session';
import bodyParser                  from 'body-parser';
import path                        from 'path';
import { fileURLToPath }           from 'url';
import { createProxyMiddleware }   from 'http-proxy-middleware';
import morgan                      from 'morgan';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const app       = express();

// 1) Middleware
app.use(morgan('dev'));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: false }));
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
app.use(express.static(path.join(__dirname, 'public')));

// 2) Proxy only /api/* → Go backend
const API_TARGET = process.env.GO_API_URL || 'http://localhost:3001';

app.use(
  '/api',
  createProxyMiddleware({
    target      : API_TARGET,
    changeOrigin: true,
    pathRewrite : { '^/api': '' },
    proxyTimeout: 65_000,
    timeout     : 70_000,
    xfwd        : true,
    onError(err, req, res) {
      console.error('Proxy error:', err);
      if (!res.headersSent) {
        res.writeHead(504, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ message: 'API gateway timeout' }));
      }
    },
  })
);

// 3) EJS setup
app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'views'));

// 4) Page routes
function auth(role) {
  return (req, res, next) => {
    if (!req.session.user) return res.redirect('/login');
    if (role && req.session.user.role !== role)
      return res.redirect(`/${req.session.user.role}`);
    next();
  };
}

const users = { alice: 'student', bob: 'instructor', iris: 'institution' };

app.get('/', (req, res) => {
  if (!req.session.user) return res.redirect('/login');
  return res.redirect(`/${req.session.user.role}`);
});

// Public sign-up page
app.get('/signup', (req, res) => {
  res.render('institution/userManagement', {
    user : null,
    title: 'Sign Up',
  });
});

app.get('/login', (req, res) =>
  res.render('login', { title: 'Log in', error: null, user: null })
);
app.post('/login', (req, res) => {
  const { username, password } = req.body;
  const role = users[username];
  if (!role || password !== '1234') {
    return res.render('login', {
      title: 'Log in',
      error: 'Invalid credentials',
      user: null,
    });
  }
  req.session.user = { username, role };
  return res.redirect(`/${role}`);
});
app.get('/logout', (req, res) =>
  req.session.destroy(() => res.redirect('/login'))
);

// Student pages
app.get('/student',            auth('student'), (r, s) => s.render('student/dashboard',      { user: r.session.user, title: 'Dashboard' }));
app.get('/student/statistics', auth('student'), (r, s) => s.render('student/statistics',     { user: r.session.user, title: 'Statistics' }));
app.get('/student/my-courses', auth('student'), (r, s) => s.render('student/myCourses',      { user: r.session.user, title: 'My Courses' }));
app.get('/student/request',    auth('student'), (r, s) => s.render('student/reviewRequest',  { user: r.session.user, title: 'Review Request' }));
app.get('/student/status',     auth('student'), (r, s) => s.render('student/reviewStatus',   { user: r.session.user, title: 'Review Status' }));
app.get('/student/personal',   auth('student'), (r, s) => s.render('student/personal',       { user: r.session.user, title: 'Personal Grades' }));

// Instructor pages
app.get('/instructor',             auth('instructor'), (r, s) => s.render('instructor/dashboard',   { user: r.session.user, title: 'Dashboard' }));
app.get('/instructor/post-initial',auth('instructor'), (r, s) => s.render('instructor/postInitial', { user: r.session.user, title: 'Post Initial' }));
app.get('/instructor/post-final',  auth('instructor'), (r, s) => s.render('instructor/postFinal',   { user: r.session.user, title: 'Post Final' }));
app.get('/instructor/review-list', auth('instructor'), (r, s) => s.render('instructor/reviewList',  { user: r.session.user, title: 'Review Requests' }));

// Reply form with template variables
app.get('/instructor/reply', auth('instructor'), (req, res) => {
  const request_id = req.query.req || '';
  res.render('instructor/replyForm', {
    user          : req.session.user,
    title         : 'Reply to Review Request',
    request_id,
    course_name   : 'software II',
    exam_period   : 'spring 2025',
    student_name  : 'john doe',
  });
});

app.get('/instructor/statistics', auth('instructor'), (r, s) => s.render('instructor/statistics',  { user: r.session.user, title: 'Statistics' }));

// Institution pages
app.get('/institution',                auth('institution'), (r, s) => s.render('institution/dashboard',       { user: r.session.user, title: 'Dashboard' }));
app.get('/institution/register',       auth('institution'), (r, s) => s.render('institution/register',        { user: r.session.user, title: 'Register' }));
app.get('/institution/purchase',       auth('institution'), (r, s) => s.render('institution/purchase',        { user: r.session.user, title: 'Purchase' }));
app.get('/institution/user-management',auth('institution'), (r, s) => s.render('institution/userManagement', { user: r.session.user, title: 'Users' }));
app.get('/institution/statistics',     auth('institution'), (r, s) => s.render('institution/statistics',      { user: r.session.user, title: 'Statistics' }));

// 5) Start server
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => console.log(`✔ Front-end listening on http://localhost:${PORT}`));